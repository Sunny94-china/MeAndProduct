package repository

import (
	"context"

	"gorm.io/gorm"

	"meandproduct/internal/model/entity"
)

type TagRepo struct {
	db *gorm.DB
}

func NewTagRepo(db *gorm.DB) *TagRepo {
	return &TagRepo{db: db}
}

func (r *TagRepo) ListActive(ctx context.Context) ([]*entity.Tag, error) {
	items := make([]*entity.Tag, 0)
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("type ASC, sort_order ASC, id ASC").
		Find(&items).Error
	return items, err
}

type GroupedTags map[entity.TagType][]*entity.Tag

func (r *TagRepo) GetByTypeGrouped(ctx context.Context, includeInactive bool) (GroupedTags, error) {
	items := make([]*entity.Tag, 0)
	q := r.db.WithContext(ctx)
	if !includeInactive {
		q = q.Where("is_active = ?", true)
	}
	q = q.Order("type ASC, sort_order ASC, id ASC")
	if err := q.Find(&items).Error; err != nil {
		return nil, err
	}
	out := GroupedTags{
		entity.TagTypeIndustry: {},
		entity.TagTypeStage:    {},
		entity.TagTypeRole:     {},
		entity.TagTypeOther:    {},
	}
	for _, t := range items {
		tp := t.Type
		if _, ok := out[tp]; !ok {
			out[entity.TagTypeOther] = append(out[entity.TagTypeOther], t)
			continue
		}
		out[tp] = append(out[tp], t)
	}
	return out, nil
}

func (r *TagRepo) Create(ctx context.Context, t *entity.Tag) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *TagRepo) Update(ctx context.Context, id uint64, fields map[string]any) error {
	if id == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&entity.Tag{}).Where("id = ?", id).Updates(fields).Error
}

func (r *TagRepo) Deactivate(ctx context.Context, id uint64) error {
	if id == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&entity.Tag{}).Where("id = ?", id).
		UpdateColumn("is_active", false).Error
}

func (r *TagRepo) GetByIDs(ctx context.Context, ids []uint64) ([]*entity.Tag, error) {
	if len(ids) == 0 {
		return []*entity.Tag{}, nil
	}
	items := make([]*entity.Tag, 0, len(ids))
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *TagRepo) IncrFounderCount(ctx context.Context, id uint64, delta int) error {
	if id == 0 || delta == 0 {
		return nil
	}
	if delta > 0 {
		return r.db.WithContext(ctx).Model(&entity.Tag{}).Where("id = ?", id).
			UpdateColumn("founder_count", gorm.Expr("founder_count + ?", uint32(delta))).Error
	}
	return r.db.WithContext(ctx).Model(&entity.Tag{}).Where("id = ?", id).
		UpdateColumn("founder_count", gorm.Expr("GREATEST(founder_count - ?, 0)", uint32(-delta))).Error
}

func (r *TagRepo) IncrProductCount(ctx context.Context, id uint64, delta int) error {
	if id == 0 || delta == 0 {
		return nil
	}
	if delta > 0 {
		return r.db.WithContext(ctx).Model(&entity.Tag{}).Where("id = ?", id).
			UpdateColumn("product_count", gorm.Expr("product_count + ?", uint32(delta))).Error
	}
	return r.db.WithContext(ctx).Model(&entity.Tag{}).Where("id = ?", id).
		UpdateColumn("product_count", gorm.Expr("GREATEST(product_count - ?, 0)", uint32(-delta))).Error
}

func (r *TagRepo) GetBySlug(ctx context.Context, slug string) (*entity.Tag, error) {
	var t entity.Tag
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).Take(&t).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}
