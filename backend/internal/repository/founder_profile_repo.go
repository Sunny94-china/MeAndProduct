package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"meandproduct/internal/model/entity"
)

type FounderProfileRepo struct {
	db *gorm.DB
}

func NewFounderProfileRepo(db *gorm.DB) *FounderProfileRepo {
	return &FounderProfileRepo{db: db}
}

type FounderListFilter struct {
	TagIDs    []uint64
	IsVisible *bool
	SortBy    string
	Page      int
	PageSize  int
}

func (r *FounderProfileRepo) GetByUserID(ctx context.Context, userID uint64) (*entity.FounderProfile, error) {
	var p entity.FounderProfile
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Take(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *FounderProfileRepo) GetBySlug(ctx context.Context, slug string, includeHidden bool) (*entity.FounderProfile, error) {
	var p entity.FounderProfile
	q := r.db.WithContext(ctx).Where("slug = ?", slug)
	if !includeHidden {
		q = q.Where("is_visible = ?", true)
	}
	if err := q.Take(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *FounderProfileRepo) Upsert(ctx context.Context, tx *gorm.DB, p *entity.FounderProfile) error {
	d := r.db
	if tx != nil {
		d = tx
	}
	return d.WithContext(ctx).Save(p).Error
}

func (r *FounderProfileRepo) Tx(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

func (r *FounderProfileRepo) TxV2(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.Tx(ctx, fn)
}

func (r *FounderProfileRepo) GetByID(ctx context.Context, id uint64) (*entity.FounderProfile, error) {
	if id == 0 {
		return nil, nil
	}
	var p entity.FounderProfile
	if err := r.db.WithContext(ctx).Where("id = ?", id).Take(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *FounderProfileRepo) GetByIDV2(ctx context.Context, id uint64) (*entity.FounderProfile, error) {
	return r.GetByID(ctx, id)
}

func (r *FounderProfileRepo) UpdateByUserID(ctx context.Context, userID uint64, fields map[string]any) error {
	return r.db.WithContext(ctx).Model(&entity.FounderProfile{}).Where("user_id = ?", userID).Updates(fields).Error
}

func (r *FounderProfileRepo) ListTags(ctx context.Context, founderID uint64) ([]*entity.Tag, error) {
	items := make([]*entity.Tag, 0)
	err := r.db.WithContext(ctx).
		Joins("JOIN founder_tags ft ON ft.tag_id = tags.id").
		Where("ft.founder_id = ?", founderID).
		Order("tags.sort_order ASC, tags.id ASC").
		Find(&items).Error
	return items, err
}

func (r *FounderProfileRepo) SetTagsReplace(ctx context.Context, tx *gorm.DB, founderID uint64, tagIDs []uint64) ([]uint64, []uint64, error) {
	d := r.db
	if tx != nil {
		d = tx
	}
	unique := make([]uint64, 0, len(tagIDs))
	seen := map[uint64]bool{}
	for _, id := range tagIDs {
		if id == 0 || seen[id] {
			continue
		}
		seen[id] = true
		unique = append(unique, id)
	}
	var existing []uint64
	if err := d.WithContext(ctx).Table("founder_tags").
		Where("founder_id = ?", founderID).Pluck("tag_id", &existing).Error; err != nil {
		return nil, nil, err
	}
	exMap := map[uint64]bool{}
	for _, id := range existing {
		exMap[id] = true
	}
	added := make([]uint64, 0)
	kept := map[uint64]bool{}
	for _, id := range unique {
		if !exMap[id] {
			added = append(added, id)
		}
		kept[id] = true
	}
	removed := make([]uint64, 0)
	for _, id := range existing {
		if !kept[id] {
			removed = append(removed, id)
		}
	}
	if len(removed) > 0 {
		if err := d.WithContext(ctx).Where("founder_id = ? AND tag_id IN ?", founderID, removed).
			Delete(&entity.FounderTag{}).Error; err != nil {
			return nil, nil, err
		}
	}
	if len(added) > 0 {
		rows := make([]entity.FounderTag, 0, len(added))
		now := time.Now()
		for _, id := range added {
			rows = append(rows, entity.FounderTag{FounderID: founderID, TagID: id, CreatedAt: now})
		}
		if err := d.WithContext(ctx).Create(&rows).Error; err != nil {
			return nil, nil, err
		}
	}
	return added, removed, nil
}

func (r *FounderProfileRepo) IncrViewCount(ctx context.Context, id uint64, delta int) error {
	if id == 0 || delta == 0 {
		return nil
	}
	if delta > 0 {
		return r.db.WithContext(ctx).Model(&entity.FounderProfile{}).Where("id = ?", id).
			UpdateColumn("view_count", gorm.Expr("view_count + ?", uint32(delta))).Error
	}
	return r.db.WithContext(ctx).Model(&entity.FounderProfile{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("GREATEST(view_count - ?, 0)", uint32(-delta))).Error
}

func (r *FounderProfileRepo) IncrFavoriteCount(ctx context.Context, id uint64, delta int) error {
	if id == 0 || delta == 0 {
		return nil
	}
	if delta > 0 {
		return r.db.WithContext(ctx).Model(&entity.FounderProfile{}).Where("id = ?", id).
			UpdateColumn("favorite_count", gorm.Expr("favorite_count + ?", uint32(delta))).Error
	}
	return r.db.WithContext(ctx).Model(&entity.FounderProfile{}).Where("id = ?", id).
		UpdateColumn("favorite_count", gorm.Expr("GREATEST(favorite_count - ?, 0)", uint32(-delta))).Error
}

func (r *FounderProfileRepo) List(ctx context.Context, filter FounderListFilter) ([]*entity.FounderProfile, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 500 {
		filter.PageSize = 20
	}
	q := r.db.WithContext(ctx).Model(&entity.FounderProfile{})
	if filter.IsVisible != nil {
		q = q.Where("is_visible = ?", *filter.IsVisible)
	}
	if len(filter.TagIDs) > 0 {
		sub := r.db.Table("founder_tags ft").Select("DISTINCT ft.founder_id").
			Where("ft.tag_id IN ?", filter.TagIDs)
		q = q.Where("id IN (?)", sub)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	switch filter.SortBy {
	case "favorite_count":
		q = q.Order("favorite_count DESC, id DESC")
	case "view_count":
		q = q.Order("view_count DESC, id DESC")
	default:
		q = q.Order("created_at DESC, id DESC")
	}
	offset := (filter.Page - 1) * filter.PageSize
	items := make([]*entity.FounderProfile, 0, filter.PageSize)
	if err := q.Offset(offset).Limit(filter.PageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

var _ = time.Now
