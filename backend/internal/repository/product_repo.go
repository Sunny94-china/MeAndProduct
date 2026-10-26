package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"meandproduct/internal/model/entity"
)

var ErrProductNotYours = errors.New("product does not belong to this founder")

type ProductRepo struct {
	db *gorm.DB
}

func NewProductRepo(db *gorm.DB) *ProductRepo {
	return &ProductRepo{db: db}
}

type ProductListFilter struct {
	FounderID *uint64
	TagIDs    []uint64
	Stage     entity.ProductStage
	IsVisible *bool
	SortBy    string
	Page      int
	PageSize  int
}

func (r *ProductRepo) GetByID(ctx context.Context, id uint64) (*entity.Product, error) {
	var p entity.Product
	if err := r.db.WithContext(ctx).Where("id = ?", id).Take(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepo) GetByFounderSlugs(ctx context.Context, founderSlug, productSlug string, includeHidden bool) (*entity.Product, error) {
	var p entity.Product
	sub := r.db.Model(&entity.FounderProfile{}).Select("id").Where("slug = ?", founderSlug)
	q := r.db.WithContext(ctx).Model(&entity.Product{}).
		Where("founder_id IN (?) AND slug = ?", sub, productSlug)
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

func (r *ProductRepo) ListByFounderID(ctx context.Context, founderID uint64, page, pageSize int) ([]*entity.Product, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 500 {
		pageSize = 50
	}
	q := r.db.WithContext(ctx).Model(&entity.Product{}).
		Where("founder_id = ?", founderID).
		Order("sort_order ASC, id DESC")
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*entity.Product, 0, pageSize)
	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *ProductRepo) Create(ctx context.Context, tx *gorm.DB, p *entity.Product) error {
	d := r.db
	if tx != nil {
		d = tx
	}
	if err := d.WithContext(ctx).Create(p).Error; err != nil {
		return err
	}
	// bool false 是零值，GORM Create 会丢掉，数据库默认值又是可见。
	if !p.IsVisible {
		return d.WithContext(ctx).Model(p).UpdateColumn("is_visible", false).Error
	}
	return nil
}

func (r *ProductRepo) Update(ctx context.Context, id uint64, fields map[string]any, checkFounderID *uint64) error {
	q := r.db.WithContext(ctx).Model(&entity.Product{}).Where("id = ?", id)
	if checkFounderID != nil {
		q = q.Where("founder_id = ?", *checkFounderID)
	}
	res := q.Updates(fields)
	if res.Error != nil {
		return res.Error
	}
	if checkFounderID != nil && res.RowsAffected == 0 {
		return ErrProductNotYours
	}
	return nil
}

func (r *ProductRepo) SoftDelete(ctx context.Context, id uint64, checkFounderID *uint64) error {
	q := r.db.WithContext(ctx).Where("id = ?", id)
	if checkFounderID != nil {
		q = q.Where("founder_id = ?", *checkFounderID)
	}
	res := q.Delete(&entity.Product{})
	if res.Error != nil {
		return res.Error
	}
	if checkFounderID != nil && res.RowsAffected == 0 {
		return ErrProductNotYours
	}
	return nil
}

func (r *ProductRepo) ListGlobal(ctx context.Context, filter ProductListFilter) ([]*entity.Product, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 500 {
		filter.PageSize = 20
	}
	q := r.db.WithContext(ctx).Model(&entity.Product{})
	if filter.FounderID != nil {
		q = q.Where("founder_id = ?", *filter.FounderID)
	}
	if filter.Stage != "" {
		q = q.Where("status = ?", filter.Stage)
	}
	if filter.IsVisible != nil {
		q = q.Where("is_visible = ?", *filter.IsVisible)
	}
	if len(filter.TagIDs) > 0 {
		sub := r.db.Table("product_tags pt").Select("DISTINCT pt.product_id").
			Where("pt.tag_id IN ?", filter.TagIDs)
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
	items := make([]*entity.Product, 0, filter.PageSize)
	if err := q.Offset(offset).Limit(filter.PageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *ProductRepo) IncrViewCount(ctx context.Context, id uint64, delta int) error {
	if id == 0 || delta == 0 {
		return nil
	}
	if delta > 0 {
		return r.db.WithContext(ctx).Model(&entity.Product{}).Where("id = ?", id).
			UpdateColumn("view_count", gorm.Expr("view_count + ?", uint32(delta))).Error
	}
	return r.db.WithContext(ctx).Model(&entity.Product{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("GREATEST(view_count - ?, 0)", uint32(-delta))).Error
}

func (r *ProductRepo) IncrFavoriteCount(ctx context.Context, id uint64, delta int) error {
	if id == 0 || delta == 0 {
		return nil
	}
	if delta > 0 {
		return r.db.WithContext(ctx).Model(&entity.Product{}).Where("id = ?", id).
			UpdateColumn("favorite_count", gorm.Expr("favorite_count + ?", uint32(delta))).Error
	}
	return r.db.WithContext(ctx).Model(&entity.Product{}).Where("id = ?", id).
		UpdateColumn("favorite_count", gorm.Expr("GREATEST(favorite_count - ?, 0)", uint32(-delta))).Error
}

func (r *ProductRepo) ListTags(ctx context.Context, productID uint64) ([]*entity.Tag, error) {
	items := make([]*entity.Tag, 0)
	err := r.db.WithContext(ctx).
		Joins("JOIN product_tags pt ON pt.tag_id = tags.id").
		Where("pt.product_id = ?", productID).
		Order("tags.sort_order ASC, tags.id ASC").
		Find(&items).Error
	return items, err
}

func (r *ProductRepo) SetTagsReplace(ctx context.Context, tx *gorm.DB, productID uint64, tagIDs []uint64) ([]uint64, []uint64, error) {
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
	if err := d.WithContext(ctx).Table("product_tags").
		Where("product_id = ?", productID).Pluck("tag_id", &existing).Error; err != nil {
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
		if err := d.WithContext(ctx).Where("product_id = ? AND tag_id IN ?", productID, removed).
			Delete(&entity.ProductTag{}).Error; err != nil {
			return nil, nil, err
		}
	}
	if len(added) > 0 {
		rows := make([]entity.ProductTag, 0, len(added))
		now := time.Now()
		for _, id := range added {
			rows = append(rows, entity.ProductTag{ProductID: productID, TagID: id, CreatedAt: now})
		}
		if err := d.WithContext(ctx).Create(&rows).Error; err != nil {
			return nil, nil, err
		}
	}
	return added, removed, nil
}

func (r *ProductRepo) Reorder(ctx context.Context, founderID uint64, orderedIDs []uint64) error {
	if len(orderedIDs) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range orderedIDs {
			if id == 0 {
				continue
			}
			res := tx.WithContext(ctx).Model(&entity.Product{}).
				Where("id = ? AND founder_id = ?", id, founderID).
				UpdateColumn("sort_order", i)
			if res.Error != nil {
				return res.Error
			}
		}
		return nil
	})
}
