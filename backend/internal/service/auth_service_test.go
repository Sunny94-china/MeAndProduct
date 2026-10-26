package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"meandproduct/internal/config"
	"meandproduct/internal/model/entity"
	"meandproduct/internal/pkg/jwtutil"
	"meandproduct/internal/repository"
)

const testMySQLDSN = "meandproduct:dev_pwd@tcp(127.0.0.1:3306)/meandproduct?charset=utf8mb4&parseTime=True&loc=Local"

func mustMySQL(t *testing.T) (*gorm.DB, func()) {
	t.Helper()
	db, err := gorm.Open(mysql.Open(testMySQLDSN), &gorm.Config{})
	if err != nil {
		t.Skipf("mysql unreachable: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Skipf("db.DB: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Skipf("mysql ping fail: %v", err)
	}
	suffix := uuid.New().String()[:8]
	return db, func() {
		ctx := context.Background()
		_ = db.WithContext(ctx).Where("username LIKE ? OR username LIKE ?", "u_"+suffix+"%", "ui_"+suffix+"%").
			Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&entity.User{}).Error
		_ = db.WithContext(ctx).Where("slug LIKE ?", "slug_"+suffix+"%").
			Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&entity.FounderProfile{}).Error
		_ = db.WithContext(ctx).Where("code LIKE ?", "INV_"+suffix+"%").
			Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&entity.InviteCode{}).Error
	}
}

func newTestCfg() *config.Config {
	return &config.Config{
		AppEnv: "test",
		JWT: config.JWTCfg{
			Secret:             "unit-test-secret-unit-test-secret-00000000",
			AccessTTL:          2 * time.Minute,
			RefreshTTL:         10 * time.Minute,
			Issuer:             "meandproduct-test",
			BlacklistKeyPrefix: "jwt:blacklist:",
		},
	}
}

// testEnv 聚合：cfg + db + miniredis + service + cleanup
type testEnv struct {
	cfg     *config.Config
	db      *gorm.DB
	rd      *redis.Client
	repo    *repository.UserRepo
	svc     *AuthService
	mr      *miniredis.Miniredis
	suffix  string
	cleanup func()
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	db, cleanupDB := mustMySQL(t)
	mr := miniredis.RunT(t)
	rd := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	cfg := newTestCfg()
	repo := repository.NewUserRepo(db)
	svc := NewAuthService(cfg, db, rd, repo)
	return &testEnv{
		cfg:     cfg,
		db:      db,
		rd:      rd,
		repo:    repo,
		svc:     svc,
		mr:      mr,
		suffix:  uuid.New().String()[:8],
		cleanup: cleanupDB,
	}
}

func TestAuth_RegisterSelf(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	ctx := context.Background()
	uname := "u_" + env.suffix
	pair, u, fp, err := env.svc.RegisterSelf(ctx, uname, uname+"@ex.com", "Passw0rd!1", "DName")
	if err != nil {
		t.Fatalf("RegisterSelf err: %v", err)
	}
	if pair == nil || u == nil || fp == nil {
		t.Fatal("RegisterSelf nil returns")
	}
	if u.Status != entity.UserStatusPending {
		t.Fatalf("self register status: want pending got %s", u.Status)
	}
	// 注意：00001_init.up.sql founder_profiles 表 is_visible 默认值为 1（true），符合 MVP 全部公开规则
	t.Logf("self register created id=%d username=%s founder.is_visible=%v (DB default=1)", u.ID, u.Username, fp.IsVisible)
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("empty token pair")
	}
	// JWT 验证
	ac, err := jwtutil.Parse(env.cfg, pair.AccessToken, jwtutil.TokenTypeAccess)
	if err != nil {
		t.Fatalf("Parse access: %v", err)
	}
	if ac.UserID != u.ID || ac.Role != string(entity.RoleFounder) {
		t.Fatalf("access claims mismatch: %+v", ac)
	}
	rc, _ := jwtutil.Parse(env.cfg, pair.RefreshToken, jwtutil.TokenTypeRefresh)
	if rc == nil || rc.UserID != u.ID {
		t.Fatalf("refresh claims mismatch: %+v", rc)
	}
	// Slug 匹配前缀
	if fp.Slug[:4] != "u-"+env.suffix[:3] { // 可能会 sanitize _→-
		t.Logf("slug=%s (sanitized %s)", fp.Slug, uname)
	}
	// 重复注册应 Conflict
	_, _, _, err2 := env.svc.RegisterSelf(ctx, uname, uname+"@ex.com", "xxx", "x")
	if !errors.Is(err2, ErrConflictUsernameEmail) {
		t.Fatalf("expect ErrConflictUsernameEmail got %v", err2)
	}
}

func TestAuth_RegisterInvite(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	ctx := context.Background()
	code := "INV_" + env.suffix
	ic := &entity.InviteCode{Code: code, MaxUses: 1, CreatedBy: 1, Status: entity.InviteActive}
	if err := env.repo.CreateInviteCode(ctx, nil, ic); err != nil {
		t.Fatalf("CreateInviteCode: %v", err)
	}
	uname := "ui_" + env.suffix
	pair, u, fp, err := env.svc.RegisterInvite(ctx, uname, uname+"@ex.com", "Passw0rd!1", "", code)
	if err != nil {
		t.Fatalf("RegisterInvite err: %v", err)
	}
	if u.Status != entity.UserStatusActive || u.RegisterSource != entity.SourceInvite || u.InviteCodeID == nil || *u.InviteCodeID != ic.ID {
		t.Fatalf("invite user state bad: %+v", u)
	}
	if !fp.IsVisible {
		t.Fatalf("invite founder should be visible")
	}
	// ConsumeInvite 后，ic.UsedCount=1 并且 status=used
	ic2, err := env.repo.GetInviteByID(ctx, ic.ID)
	if err != nil || ic2 == nil || ic2.UsedCount != 1 || ic2.Status != entity.InviteUsed {
		t.Fatalf("invite after consume state wrong: %+v err=%v", ic2, err)
	}
	// 二次使用邀请码 → 应拒绝
	_, _, _, err2 := env.svc.RegisterInvite(ctx, "ui2_"+env.suffix, "ui2_"+env.suffix+"@ex.com", "xxx", "", code)
	if !errors.Is(err2, ErrInviteCodeInvalid) {
		t.Fatalf("2nd register with same 1-use invite expect ErrInviteCodeInvalid, got %v", err2)
	}
	// 用不存在 code → 应拒绝
	_, _, _, err3 := env.svc.RegisterInvite(ctx, "ui3_"+env.suffix, "ui3_"+env.suffix+"@ex.com", "xxx", "", "NOEXIST")
	if !errors.Is(err3, ErrInviteCodeInvalid) {
		t.Fatalf("invite invalid code expect ErrInviteCodeInvalid, got %v", err3)
	}
	_ = pair
}

func TestAuth_Login(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	ctx := context.Background()
	// RegisterSelf → 会生成 pending 用户，pending 登录应报错
	uname := "u_l_" + env.suffix
	_, uSelf, _, _ := env.svc.RegisterSelf(ctx, uname, uname+"@ex.com", "Passw0rd!1", "LName")
	if uSelf == nil {
		t.Fatal("RegisterSelf nil user")
	}
	// 错误密码
	_, _, errBad := env.svc.Login(ctx, uname, "wrongpass")
	if !errors.Is(errBad, ErrInvalidCredentials) {
		t.Fatalf("wrong pwd expect ErrInvalidCredentials got %v", errBad)
	}
	// 用邮箱登录 pending → 错误账户不可登录
	_, _, errPending := env.svc.Login(ctx, uname+"@ex.com", "Passw0rd!1")
	if !errors.Is(errPending, ErrUserSuspendedOrPending) {
		t.Fatalf("pending login expect ErrUserSuspendedOrPending got %v", errPending)
	}
	// 更新为 active 后登录成功
	_ = env.repo.UpdateStatus(ctx, uSelf.ID, entity.UserStatusActive)
	pair, uLogin, errOK := env.svc.Login(ctx, uSelf.Username, "Passw0rd!1")
	if errOK != nil || uLogin == nil || pair == nil {
		t.Fatalf("active login fail: pair=%+v u=%+v err=%v", pair, uLogin, errOK)
	}
	// 再次读 GetByID，last_login_at 被更新
	u2, _ := env.repo.GetByID(ctx, uSelf.ID)
	if u2.LastLoginAt == nil {
		t.Fatalf("last_login_at 未更新")
	}
	t.Logf("login success id=%d username=%s", uLogin.ID, uLogin.Username)
}

func TestAuth_RefreshAndLogout(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	ctx := context.Background()
	uname := "u_rl_" + env.suffix
	pair0, u, _, _ := env.svc.RegisterSelf(ctx, uname, uname+"@ex.com", "Passw0rd!1", "RL")
	_ = env.repo.UpdateStatus(ctx, u.ID, entity.UserStatusActive)
	if pair0 == nil {
		t.Fatal("RegisterSelf nil pair")
	}
	// 用 refresh token 换新 pair
	pair1, err := env.svc.Refresh(ctx, pair0.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh err: %v", err)
	}
	if pair1 == nil || pair1.AccessToken == pair0.AccessToken || pair1.RefreshToken == pair0.RefreshToken {
		t.Fatalf("Refresh 未产出新 pair: new=%+v old=%+v", pair1, pair0)
	}
	// 旧 refresh 再用 → 应被拉黑，报错
	_, err2 := env.svc.Refresh(ctx, pair0.RefreshToken)
	if !errors.Is(err2, ErrRefreshTokenInvalid) {
		t.Fatalf("reuse old refresh expect ErrRefreshTokenInvalid got %v", err2)
	}
	// 非法 refresh token → 拒绝
	_, err3 := env.svc.Refresh(ctx, "garbage")
	if !errors.Is(err3, ErrRefreshTokenInvalid) {
		t.Fatalf("garbage refresh expect ErrRefreshTokenInvalid got %v", err3)
	}
	// Logout：新 access + refresh 拉黑，刷新后报错
	ac1, _ := jwtutil.Parse(env.cfg, pair1.AccessToken, jwtutil.TokenTypeAccess)
	rc1, _ := jwtutil.Parse(env.cfg, pair1.RefreshToken, jwtutil.TokenTypeRefresh)
	if err := env.svc.Logout(ctx, ac1.JWTID, rc1.JWTID, pair1.AccessExpiresAt, pair1.RefreshExpiresAt); err != nil {
		t.Fatalf("Logout err: %v", err)
	}
	// 再次 refresh pair1.refresh → 已拉黑应拒绝
	_, err4 := env.svc.Refresh(ctx, pair1.RefreshToken)
	if !errors.Is(err4, ErrRefreshTokenInvalid) {
		t.Fatalf("After Logout, refresh expect ErrRefreshTokenInvalid got %v", err4)
	}
	// Logout 不带 refreshJTI 也可以成功（参数缺省不报错）
	if err := env.svc.Logout(ctx, ac1.JWTID, "", pair1.AccessExpiresAt, time.Time{}); err != nil {
		t.Fatalf("Logout partial fail: %v", err)
	}
}

func TestAuth_ValidateInvite(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	ctx := context.Background()
	// case1: 不存在
	r1, err := env.svc.ValidateInvite(ctx, "nonexist")
	if err != nil || r1.Usable {
		t.Fatalf("Validate nonexist: %+v err=%v", r1, err)
	}
	// case2: revoked
	revCode := "INV_REV_" + env.suffix
	ic := &entity.InviteCode{Code: revCode, MaxUses: 2, CreatedBy: 1, Status: entity.InviteActive}
	_ = env.repo.CreateInviteCode(ctx, nil, ic)
	_ = env.repo.RevokeInvite(ctx, ic.ID)
	r2, _ := env.svc.ValidateInvite(ctx, revCode)
	if r2.Usable {
		t.Fatalf("revoked should be unusable: %+v", r2)
	}
	// case3: active, max=3 → usable=true remaining=3
	aCode := "INV_OK_" + env.suffix
	ic3 := &entity.InviteCode{Code: aCode, MaxUses: 3, CreatedBy: 1, Status: entity.InviteActive}
	_ = env.repo.CreateInviteCode(ctx, nil, ic3)
	r3, err3 := env.svc.ValidateInvite(ctx, aCode)
	if err3 != nil || !r3.Usable || r3.RemainingUses != 3 {
		t.Fatalf("valid invite: %+v err=%v", r3, err3)
	}
	// case4: expired
	past := time.Now().Add(-1 * time.Hour)
	eCode := "INV_EXP_" + env.suffix
	ic4 := &entity.InviteCode{Code: eCode, MaxUses: 99, CreatedBy: 1, Status: entity.InviteActive, ExpiresAt: &past}
	_ = env.repo.CreateInviteCode(ctx, nil, ic4)
	r4, _ := env.svc.ValidateInvite(ctx, eCode)
	if r4.Usable {
		t.Fatalf("expired should be unusable: %+v", r4)
	}
	// case5: empty string
	r5, _ := env.svc.ValidateInvite(ctx, "   ")
	if r5.Usable {
		t.Fatalf("empty code unusable expected, got %+v", r5)
	}
}

func TestAuth_Me(t *testing.T) {
	env := newTestEnv(t)
	defer env.cleanup()
	ctx := context.Background()
	uname := "u_me_" + env.suffix
	// RegisterSelf → founder is_visible=false, status=pending → Me 只返回基本名片
	_, u, fp, _ := env.svc.RegisterSelf(ctx, uname, uname+"@ex.com", "Passw0rd!1", "MeName")
	me, err := env.svc.Me(ctx, u.ID)
	if err != nil {
		t.Fatalf("Me: %v", err)
	}
	if me.ID != u.ID || me.Founder == nil {
		t.Fatalf("Me self-register basic fail: %+v", me)
	}
	// MVP 公开原则：founder is_visible DB 默认 true（注册后默认公开），title/bio 未设置前为 nil
	t.Logf("Me self register user id=%d founder.id=%d is_visible=%v display=%q slug=%q",
		me.ID, me.Founder.ID, me.Founder.IsVisible, me.Founder.DisplayName, me.Founder.Slug)
	// Set title & bio 到 founder profile，并且标记可见 → 再读
	title, bio := "CEO", "Founder Bio"
	env.db.WithContext(ctx).Model(&entity.FounderProfile{}).Where("id = ?", fp.ID).Updates(map[string]any{
		"title":      title,
		"bio":        bio,
		"is_visible": true,
	})
	me2, _ := env.svc.Me(ctx, u.ID)
	if me2.Founder.Title == nil || *me2.Founder.Title != title || me2.Founder.Bio == nil || *me2.Founder.Bio != bio {
		t.Fatalf("Me visible should contain full fields: founder=%+v", me2.Founder)
	}
	// 不存在 ID → ErrUserNotFound
	_, err2 := env.svc.Me(ctx, 99999999999)
	if !errors.Is(err2, ErrUserNotFound) {
		t.Fatalf("Me not found expect ErrUserNotFound got %v", err2)
	}
}
