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

func newTestDBP3(t *testing.T) (*gorm.DB, func()) {
	t.Helper()
	mustMySQL(t)
	db, err := gorm.Open(mysql.Open(testDSN), &gorm.Config{Logger: gl.Default.LogMode(gl.Warn)})
	if err != nil {
		t.Fatalf("connect mysql fail: %v", err)
	}
	suffix := uuid.New().String()[:8]
	ctx := context.Background()
	return db, func() {
		_ = db.WithContext(ctx).Where("name LIKE ? OR slug LIKE ?", "t3_"+suffix+"%", "t3_"+suffix+"%").
			Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&entity.Tag{}).Error
		_ = db.WithContext(ctx).Where("name LIKE ? OR slug LIKE ?", "p3_"+suffix+"%", "p3_"+suffix+"%").
			Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&entity.Product{}).Error
		_ = db.WithContext(ctx).Where("slug LIKE ?", "f3_"+suffix+"%").
			Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&entity.FounderProfile{}).Error
		_ = db.WithContext(ctx).Where("username LIKE ? OR email LIKE ?", "u3_"+suffix+"%", "u3_"+suffix+"%").
			Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&entity.User{}).Error
		_ = db.WithContext(ctx).Where("target_type IN ?", []entity.FavoriteTarget{
			entity.FavoriteTargetFounder, entity.FavoriteTargetProduct,
		}).Where("user_id IN (SELECT id FROM users WHERE username LIKE ?)", "u3_"+suffix+"%").
			Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&entity.Favorite{}).Error
		_ = db.WithContext(ctx).Where("target_type IN ?", []entity.ReportTarget{
			entity.ReportTargetFounder, entity.ReportTargetProduct,
		}).Where("reason_detail LIKE ?", "r3_"+suffix+"%").
			Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&entity.Report{}).Error
		_ = db.WithContext(ctx).Where("source_type IN ?", []entity.LinkSource{
			entity.LinkSourceFounder, entity.LinkSourceProduct,
		}).Where("link_type LIKE ?", "l3_"+suffix+"%").
			Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&entity.Link{}).Error
	}
}

// 必须创建一个 user + founder_profile（因为 founder_id = user_id，和现有 schema 一致）
func createTestFounderProfile(ctx context.Context, t *testing.T, db *gorm.DB, suffix string) (*entity.User, *entity.FounderProfile) {
	t.Helper()
	uname := "u3_" + suffix
	u := &entity.User{
		Username:       uname,
		Email:          uname + "@x.com",
		PasswordHash:   "x",
		Role:           entity.RoleFounder,
		Status:         entity.UserStatusActive,
		RegisterSource: entity.SourceSelf,
	}
	if err := db.WithContext(ctx).Create(u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	fp := &entity.FounderProfile{
		UserID:      u.ID,
		DisplayName: "Founder " + suffix,
		Slug:        "f3_" + suffix,
		IsVisible:   true,
	}
	if err := db.WithContext(ctx).Create(fp).Error; err != nil {
		t.Fatalf("create founder profile: %v", err)
	}
	return u, fp
}

func TestP3Product_SlugUniq(t *testing.T) {
	db, cleanup := newTestDBP3(t)
	defer cleanup()
	ctx := context.Background()
	suffix := uuid.New().String()[:8]
	_, fp := createTestFounderProfile(ctx, t, db, suffix)

	repo := NewProductRepo(db)
	p1 := &entity.Product{
		FounderID:   fp.ID,
		Name:        "p3_" + suffix + "_prod1",
		Slug:        "p3_" + suffix + "_slug1",
		Tagline:     "tag line 1",
		OfficialURL: "https://example.com/1",
		Status:      entity.StageBeta,
		IsVisible:   true,
	}
	if err := repo.Create(ctx, nil, p1); err != nil {
		t.Fatalf("Create p1: %v", err)
	}
	if p1.ID == 0 {
		t.Fatalf("Create p1 id=0")
	}
	p2 := &entity.Product{
		FounderID:   fp.ID,
		Name:        "p3_" + suffix + "_prod2",
		Slug:        "p3_" + suffix + "_slug1",
		Tagline:     "tag line 2",
		OfficialURL: "https://example.com/2",
		Status:      entity.StageIdea,
		IsVisible:   true,
	}
	if err := repo.Create(ctx, nil, p2); err == nil {
		t.Fatalf("expected dup slug error, got nil")
	}

	got, err := repo.GetByFounderSlugs(ctx, fp.Slug, p1.Slug, true)
	if err != nil || got == nil || got.ID != p1.ID {
		t.Fatalf("GetByFounderSlugs: err=%v got=%+v", err, got)
	}
}

func TestP3Favorite_Idempotent(t *testing.T) {
	db, cleanup := newTestDBP3(t)
	defer cleanup()
	ctx := context.Background()
	suffix := uuid.New().String()[:8]
	u, fp := createTestFounderProfile(ctx, t, db, suffix)

	favRepo := NewFavoriteRepo(db)
	existed, err := favRepo.Upsert(ctx, u.ID, entity.FavoriteTargetFounder, fp.ID)
	if err != nil || existed {
		t.Fatalf("Upsert #1: existed=%v err=%v", existed, err)
	}
	existed2, err := favRepo.Upsert(ctx, u.ID, entity.FavoriteTargetFounder, fp.ID)
	if err != nil || !existed2 {
		t.Fatalf("Upsert #2（幂等）: existed=%v err=%v", existed2, err)
	}
	ok, err := favRepo.Exists(ctx, u.ID, entity.FavoriteTargetFounder, fp.ID)
	if err != nil || !ok {
		t.Fatalf("Exists: ok=%v err=%v", ok, err)
	}
	n, err := favRepo.Delete(ctx, u.ID, entity.FavoriteTargetFounder, fp.ID)
	if err != nil || n != 1 {
		t.Fatalf("Delete #1: n=%d err=%v", n, err)
	}
	n2, err := favRepo.Delete(ctx, u.ID, entity.FavoriteTargetFounder, fp.ID)
	if err != nil || n2 != 0 {
		t.Fatalf("Delete #2（空删）: n=%d err=%v", n2, err)
	}
	ok2, _ := favRepo.Exists(ctx, u.ID, entity.FavoriteTargetFounder, fp.ID)
	if ok2 {
		t.Fatalf("after delete still exists")
	}
}

func TestP3Report_Handle(t *testing.T) {
	db, cleanup := newTestDBP3(t)
	defer cleanup()
	ctx := context.Background()
	suffix := uuid.New().String()[:8]
	_, fp := createTestFounderProfile(ctx, t, db, suffix)

	rpRepo := NewReportRepo(db)
	rp := &entity.Report{
		TargetType:   entity.ReportTargetFounder,
		TargetID:     fp.ID,
		ReasonType:   entity.ReasonSpam,
		ReasonDetail: "r3_" + suffix + " spam spam spam",
	}
	if err := rpRepo.Create(ctx, nil, rp); err != nil {
		t.Fatalf("Create report: %v", err)
	}
	if rp.ID == 0 {
		t.Fatalf("report id=0")
	}
	note := "handled: hide target"
	tt, tID, err := rpRepo.Handle(ctx, nil, rp.ID, HandleReportPayload{
		HandlerID:   999,
		NewStatus:   entity.ReportResolved,
		HandlerNote: &note,
		HideTarget:  true,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if tt != entity.ReportTargetFounder || tID != fp.ID {
		t.Fatalf("Handle return target mismatch: tt=%v tid=%d", tt, tID)
	}
	got, _ := rpRepo.GetByID(ctx, rp.ID)
	if got == nil || got.Status != entity.ReportResolved || got.HandlerNote == nil || *got.HandlerNote != note {
		t.Fatalf("after handle: got=%+v", got)
	}
	fpRepo := NewFounderProfileRepo(db)
	fpAfter, err := fpRepo.GetBySlug(ctx, fp.Slug, true)
	if err != nil || fpAfter == nil || fpAfter.IsVisible {
		t.Fatalf("Handle hide target failed: fpAfter=%+v err=%v", fpAfter, err)
	}
}

func TestP3Tag_Deactivate(t *testing.T) {
	db, cleanup := newTestDBP3(t)
	defer cleanup()
	ctx := context.Background()
	suffix := uuid.New().String()[:8]

	repo := NewTagRepo(db)
	tags := []*entity.Tag{
		{Name: "t3_" + suffix + "ind1", Slug: "t3_" + suffix + "slug1", Type: entity.TagTypeIndustry, IsActive: true},
		{Name: "t3_" + suffix + "ind2", Slug: "t3_" + suffix + "slug2", Type: entity.TagTypeIndustry, IsActive: true},
		{Name: "t3_" + suffix + "st1", Slug: "t3_" + suffix + "slug3", Type: entity.TagTypeStage, IsActive: true},
	}
	for _, tg := range tags {
		if err := repo.Create(ctx, tg); err != nil {
			t.Fatalf("create tag: %v", err)
		}
	}
	activeList, err := repo.ListActive(ctx)
	if err != nil {
		t.Fatalf("ListActive: %v", err)
	}
	before := 0
	for _, tg := range activeList {
		if tg.Name == "t3_"+suffix+"ind1" || tg.Name == "t3_"+suffix+"ind2" || tg.Name == "t3_"+suffix+"st1" {
			before++
		}
	}
	if before != 3 {
		t.Fatalf("before deactivate active count=%d want=3 list=%+v", before, activeList)
	}
	if err := repo.Deactivate(ctx, tags[1].ID); err != nil {
		t.Fatalf("Deactivate: %v", err)
	}
	after, err := repo.ListActive(ctx)
	if err != nil {
		t.Fatalf("ListActive after: %v", err)
	}
	afterCnt := 0
	for _, tg := range after {
		if tg.Name == "t3_"+suffix+"ind1" || tg.Name == "t3_"+suffix+"ind2" || tg.Name == "t3_"+suffix+"st1" {
			afterCnt++
		}
	}
	if afterCnt != 2 {
		t.Fatalf("after deactivate active count=%d want=2 list=%+v", afterCnt, after)
	}
	grouped, err := repo.GetByTypeGrouped(ctx, false)
	if err != nil {
		t.Fatalf("Grouped: %v", err)
	}
	var inCnt, stCnt int
	for _, tg := range grouped[entity.TagTypeIndustry] {
		if tg.Name == "t3_"+suffix+"ind1" || tg.Name == "t3_"+suffix+"ind2" {
			inCnt++
		}
	}
	for _, tg := range grouped[entity.TagTypeStage] {
		if tg.Name == "t3_"+suffix+"st1" {
			stCnt++
		}
	}
	if inCnt != 1 {
		t.Fatalf("grouped industry (ours)=%d want=1", inCnt)
	}
	if stCnt != 1 {
		t.Fatalf("grouped stage (ours)=%d want=1", stCnt)
	}
}

func TestP3Link_UpsertListIncr(t *testing.T) {
	db, cleanup := newTestDBP3(t)
	defer cleanup()
	ctx := context.Background()
	suffix := uuid.New().String()[:8]
	_, fp := createTestFounderProfile(ctx, t, db, suffix)

	repo := NewLinkRepo(db)
	rows := []LinkRow{
		{LinkType: "l3_" + suffix + "_gh", URL: "https://github.com/" + suffix},
		{LinkType: "l3_" + suffix + "_tw", URL: "https://x.com/" + suffix},
		{LinkType: "l3_" + suffix + "_tw", URL: "https://x.com/" + suffix},
	}
	if err := repo.BatchUpsert(ctx, nil, entity.LinkSourceFounder, fp.ID, rows); err != nil {
		t.Fatalf("BatchUpsert: %v", err)
	}
	list, err := repo.ListBySource(ctx, entity.LinkSourceFounder, fp.ID)
	if err != nil || len(list) != 2 {
		t.Fatalf("ListBySource len=%d want=2 err=%v list=%+v", len(list), err, list)
	}
	affected, err := repo.IncrClick(ctx, list[0].ID, 3)
	if err != nil || affected != 1 {
		t.Fatalf("IncrClick: affected=%d err=%v", affected, err)
	}
	again, _ := repo.ListBySource(ctx, entity.LinkSourceFounder, fp.ID)
	if again[0].ClickCount != 3 {
		t.Fatalf("click count=%d want=3", again[0].ClickCount)
	}
}

var _ = time.Now
