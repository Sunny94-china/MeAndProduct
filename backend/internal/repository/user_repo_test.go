package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gl "gorm.io/gorm/logger"

	"meandproduct/internal/model/entity"
)

// DSN 使用独立库 meandproduct_test（测试后清理）
const testDSN = "meandproduct:dev_pwd@tcp(127.0.0.1:3306)/meandproduct?charset=utf8mb4&parseTime=True&loc=Local"

func newTestDB(t *testing.T) (*gorm.DB, func()) {
	t.Helper()
	db, err := gorm.Open(mysql.Open(testDSN), &gorm.Config{Logger: gl.Default.LogMode(gl.Warn)})
	if err != nil {
		t.Fatalf("connect mysql fail: %v", err)
	}
	// 先删测试行（用软删保留约束）
	suffix := uuid.New().String()[:8]
	ctx := context.Background()
	return db, func() {
		// 倒序清理：invite_codes → founder_profiles → users
		_ = db.WithContext(ctx).Where("username LIKE ? OR username LIKE ?", "tu_"+suffix+"%", "user"+suffix+"%").
			Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&entity.User{}).Error
		_ = db.WithContext(ctx).Where("slug LIKE ?", "slug"+suffix+"%").
			Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&entity.FounderProfile{}).Error
		_ = db.WithContext(ctx).Where("code LIKE ?", "INV"+suffix+"%").
			Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&entity.InviteCode{}).Error
	}
}

// 保证测试 DB 可达（非 MySQL 下 t.Skip）
func mustMySQL(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(mysql.Open(testDSN), &gorm.Config{})
	if err != nil {
		t.Skipf("mysql unreachable, skip: %v", err)
	}
	sqlDB, _ := db.DB()
	if err := sqlDB.Ping(); err != nil {
		t.Skipf("mysql ping fail, skip: %v", err)
	}
}

func TestUser_CRUD(t *testing.T) {
	mustMySQL(t)
	db, cleanup := newTestDB(t)
	defer cleanup()
	repo := NewUserRepo(db)
	ctx := context.Background()
	suffix := uuid.New().String()[:8]
	uname := "tu_" + suffix

	u := &entity.User{
		Username:       uname,
		Email:          uname + "@example.com",
		PasswordHash:   "hash_x",
		Role:           entity.RoleFounder,
		Status:         entity.UserStatusActive,
		RegisterSource: entity.SourceSelf,
	}
	if err := repo.CreateUser(ctx, nil, u); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if u.ID == 0 {
		t.Fatal("CreateUser: id not filled")
	}
	got, err := repo.GetByUsername(ctx, uname)
	if err != nil || got == nil || got.ID != u.ID {
		t.Fatalf("GetByUsername: %v got=%+v", err, got)
	}
	exists, err := repo.UsernameOrEmailExists(ctx, uname, "none"+suffix+"@x.com")
	if err != nil || !exists {
		t.Fatalf("UsernameOrEmailExists(user): %v %v", err, exists)
	}
	exists2, _ := repo.UsernameOrEmailExists(ctx, "no"+suffix, uname+"@example.com")
	if !exists2 {
		t.Fatalf("UsernameOrEmailExists(email): false")
	}
	now := time.Now()
	if err := repo.UpdateLastLoginAt(ctx, u.ID, now); err != nil {
		t.Fatalf("UpdateLastLoginAt: %v", err)
	}
	got2, _ := repo.GetByID(ctx, u.ID)
	if got2.LastLoginAt == nil {
		t.Fatalf("UpdateLastLoginAt not persisted: last_login=nil")
	}
	// MySQL DATETIME 精度：秒级对齐
	if !got2.LastLoginAt.After(now.Add(-2*time.Second)) || !got2.LastLoginAt.Before(now.Add(2*time.Second)) {
		t.Fatalf("UpdateLastLoginAt out of range: last_login=%v ref_now=%v", got2.LastLoginAt, now)
	}
	if err := repo.UpdateStatus(ctx, u.ID, entity.UserStatusSuspended); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	got3, _ := repo.GetByID(ctx, u.ID)
	if got3.Status != entity.UserStatusSuspended {
		t.Fatalf("UpdateStatus mismatch: %s", got3.Status)
	}
	gotByEmail, err := repo.GetByEmail(ctx, u.Email)
	if err != nil || gotByEmail == nil || gotByEmail.ID != u.ID {
		t.Fatalf("GetByEmail: %v got=%+v", err, gotByEmail)
	}
}

func TestFounderProfile_CreateAndSlug(t *testing.T) {
	mustMySQL(t)
	db, cleanup := newTestDB(t)
	defer cleanup()
	repo := NewUserRepo(db)
	ctx := context.Background()
	suffix := uuid.New().String()[:8]

	u := &entity.User{
		Username:       "user" + suffix,
		Email:          "user" + suffix + "@example.com",
		PasswordHash:   "z",
		Role:           entity.RoleFounder,
		Status:         entity.UserStatusActive,
		RegisterSource: entity.SourceSelf,
	}
	if err := repo.CreateUser(ctx, nil, u); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	slug := "slug" + suffix
	p := &entity.FounderProfile{
		UserID:      u.ID,
		DisplayName: "DName",
		Slug:        slug,
		IsVisible:   true,
	}
	if err := repo.CreateFounderProfile(ctx, nil, p); err != nil {
		t.Fatalf("CreateFounderProfile: %v", err)
	}
	ok, _ := repo.FounderSlugExists(ctx, slug)
	if !ok {
		t.Fatalf("FounderSlugExists want true, got false")
	}
	ok2, _ := repo.FounderSlugExists(ctx, "never"+suffix)
	if ok2 {
		t.Fatalf("FounderSlugExists want false, got true")
	}
	got, err := repo.GetFounderProfileByUserID(ctx, u.ID)
	if err != nil || got == nil || got.Slug != slug {
		t.Fatalf("GetFounderProfileByUserID: %v got=%+v", err, got)
	}
}

func TestInviteCode_CreateListConsumeRevoke(t *testing.T) {
	mustMySQL(t)
	db, cleanup := newTestDB(t)
	defer cleanup()
	repo := NewUserRepo(db)
	ctx := context.Background()
	suffix := uuid.New().String()[:8]
	code := "INV" + suffix

	ic := &entity.InviteCode{
		Code:      code,
		MaxUses:   2,
		UsedCount: 0,
		Status:    entity.InviteActive,
		ExpiresAt: nil,
		CreatedBy: 1,
	}
	if err := repo.CreateInviteCode(ctx, nil, ic); err != nil {
		t.Fatalf("CreateInviteCode: %v", err)
	}
	got, err := repo.GetInviteByCode(ctx, code)
	if err != nil || got == nil || got.ID != ic.ID {
		t.Fatalf("GetInviteByCode: %v got=%+v", err, got)
	}
	if !got.Usable(time.Now()) {
		t.Fatalf("Usable want true, got false: %+v", got)
	}
	list, total, err := repo.ListInvitesByCreator(ctx, 1, 100, 0)
	if err != nil || total < 1 {
		t.Fatalf("ListInvitesByCreator: total=%d err=%v", total, err)
	}
	// 找到刚才创建的
	found := false
	for _, x := range list {
		if x.Code == code {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("ListInvitesByCreator 未找到 code=%s list=%d", code, len(list))
	}

	now := time.Now()
	// 第一次 consume
	n, err := repo.ConsumeInvite(ctx, nil, ic.ID, now)
	if err != nil || n != 1 {
		t.Fatalf("ConsumeInvite #1: n=%d err=%v", n, err)
	}
	got2, _ := repo.GetInviteByID(ctx, ic.ID)
	if got2.UsedCount != 1 || got2.Status != entity.InviteActive {
		t.Fatalf("Consume #1 after state: used=%d status=%s", got2.UsedCount, got2.Status)
	}
	// 第二次 consume 触发 max_uses → status=used
	n2, err := repo.ConsumeInvite(ctx, nil, ic.ID, now)
	if err != nil || n2 != 1 {
		t.Fatalf("ConsumeInvite #2: n=%d err=%v", n2, err)
	}
	got3, _ := repo.GetInviteByID(ctx, ic.ID)
	if got3.UsedCount != 2 || got3.Status != entity.InviteUsed {
		t.Fatalf("Consume #2 expect used=2 status=used, got used=%d status=%s", got3.UsedCount, got3.Status)
	}
	// 第三次应 0（不可用）
	n3, err := repo.ConsumeInvite(ctx, nil, ic.ID, now)
	if err != nil || n3 != 0 {
		t.Fatalf("ConsumeInvite #3 expect 0, got n=%d err=%v", n3, err)
	}

	// Revoke: 新建一个 active 的再 revoke
	ic2 := &entity.InviteCode{
		Code:      code + "_rev",
		MaxUses:   9,
		Status:    entity.InviteActive,
		CreatedBy: 1,
	}
	_ = repo.CreateInviteCode(ctx, nil, ic2)
	if err := repo.RevokeInvite(ctx, ic2.ID); err != nil {
		t.Fatalf("RevokeInvite: %v", err)
	}
	gotR, _ := repo.GetInviteByID(ctx, ic2.ID)
	if gotR.Status != entity.InviteRevoked {
		t.Fatalf("RevokeInvite 后 status 不对: %s", gotR.Status)
	}
	if gotR.Usable(time.Now()) {
		t.Fatalf("Revoked Usable=true 错误")
	}

	// MarkExpired: 建一个已过期 invite
	exp := time.Now().Add(-1 * time.Hour)
	ic3 := &entity.InviteCode{
		Code:      code + "_exp",
		MaxUses:   9,
		Status:    entity.InviteActive,
		ExpiresAt: &exp,
		CreatedBy: 1,
	}
	_ = repo.CreateInviteCode(ctx, nil, ic3)
	markedN, err := repo.MarkInviteExpired(ctx, time.Now())
	if err != nil {
		t.Fatalf("MarkInviteExpired: %v", err)
	}
	if markedN < 1 {
		t.Fatalf("MarkInviteExpired 预期 >=1, got %d", markedN)
	}
	gotE, _ := repo.GetInviteByID(ctx, ic3.ID)
	if gotE.Status != entity.InviteExpired {
		t.Fatalf("MarkExpired 后状态应为 expired, got=%s", gotE.Status)
	}
}
