package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"meandproduct/internal/model/entity"
)

type LinkRow struct {
	LinkType string
	URL      string
}

type LinkRepo struct {
	db *gorm.DB
}

func NewLinkRepo(db *gorm.DB) *LinkRepo {
	return &LinkRepo{db: db}
}

func validSource(t entity.LinkSource) bool {
	return t == entity.LinkSourceFounder || t == entity.LinkSourceProduct
}

func (r *LinkRepo) ListBySource(ctx context.Context, sourceType entity.LinkSource, sourceID uint64) ([]*entity.Link, error) {
	items := make([]*entity.Link, 0)
	if !validSource(sourceType) || sourceID == 0 {
		return items, nil
	}
	err := r.db.WithContext(ctx).
		Where("source_type = ? AND source_id = ?", sourceType, sourceID).
		Order("id ASC").Find(&items).Error
	return items, err
}

func (r *LinkRepo) BatchUpsert(ctx context.Context, tx *gorm.DB, sourceType entity.LinkSource, sourceID uint64, rows []LinkRow) error {
	d := r.db
	if tx != nil {
		d = tx
	}
	if !validSource(sourceType) || sourceID == 0 {
		return nil
	}
	if err := d.WithContext(ctx).
		Where("source_type = ? AND source_id = ?", sourceType, sourceID).
		Delete(&entity.Link{}).Error; err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	now := time.Now()
	entities := make([]entity.Link, 0, len(rows))
	seen := map[[2]string]bool{}
	for _, row := range rows {
		if row.URL == "" || row.LinkType == "" {
			continue
		}
		key := [2]string{row.LinkType, row.URL}
		if seen[key] {
			continue
		}
		seen[key] = true
		entities = append(entities, entity.Link{
			SourceType: sourceType,
			SourceID:   sourceID,
			LinkType:   row.LinkType,
			URL:        row.URL,
			ClickCount: 0,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}
	if len(entities) == 0 {
		return nil
	}
	return d.WithContext(ctx).Create(&entities).Error
}

func (r *LinkRepo) IncrClick(ctx context.Context, id uint64, delta int) (int64, error) {
	if id == 0 || delta == 0 {
		return 0, nil
	}
	res := r.db.WithContext(ctx).Model(&entity.Link{}).
		Where("id = ?", id).
		UpdateColumn("click_count", gorm.Expr("click_count + ?", delta))
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

func (r *LinkRepo) GetByID(ctx context.Context, id uint64) (*entity.Link, error) {
	if id == 0 {
		return nil, nil
	}
	var l entity.Link
	if err := r.db.WithContext(ctx).Where("id = ?", id).Take(&l).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &l, nil
}
