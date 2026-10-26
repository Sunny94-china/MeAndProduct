package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"meandproduct/internal/model/entity"
)

var ErrInvalidFavoriteTarget = errors.New("invalid favorite target_type")

type FavoriteRepo struct {
	db *gorm.DB
}

func NewFavoriteRepo(db *gorm.DB) *FavoriteRepo {
	return &FavoriteRepo{db: db}
}

type FavoriteListFilter struct {
	TargetType entity.FavoriteTarget
	Page       int
	PageSize   int
}

func validTarget(t entity.FavoriteTarget) bool {
	return entity.ValidFavoriteTargets[t]
}

func (r *FavoriteRepo) Upsert(ctx context.Context, userID uint64, targetType entity.FavoriteTarget, targetID uint64) (existed bool, err error) {
	if userID == 0 || targetID == 0 || !validTarget(targetType) {
		return false, ErrInvalidFavoriteTarget
	}
	f := &entity.Favorite{
		UserID:     userID,
		TargetType: targetType,
		TargetID:   targetID,
	}
	res := r.db.WithContext(ctx).Where("user_id = ? AND target_type = ? AND target_id = ?",
		userID, targetType, targetID).
		First(&entity.Favorite{})
	if res.Error == nil {
		return true, nil
	}
	if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return false, res.Error
	}
	if err := r.db.WithContext(ctx).Create(f).Error; err != nil {
		return false, err
	}
	return false, nil
}

func (r *FavoriteRepo) Delete(ctx context.Context, userID uint64, targetType entity.FavoriteTarget, targetID uint64) (int64, error) {
	if userID == 0 || targetID == 0 || !validTarget(targetType) {
		return 0, ErrInvalidFavoriteTarget
	}
	res := r.db.WithContext(ctx).
		Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
		Delete(&entity.Favorite{})
	return res.RowsAffected, res.Error
}

func (r *FavoriteRepo) Exists(ctx context.Context, userID uint64, targetType entity.FavoriteTarget, targetID uint64) (bool, error) {
	if userID == 0 || targetID == 0 || !validTarget(targetType) {
		return false, nil
	}
	var c int64
	err := r.db.WithContext(ctx).Model(&entity.Favorite{}).
		Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
		Count(&c).Error
	return c > 0, err
}

func (r *FavoriteRepo) ListByUser(ctx context.Context, userID uint64, filter FavoriteListFilter) ([]*entity.Favorite, int64, error) {
	if userID == 0 {
		return []*entity.Favorite{}, 0, nil
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 500 {
		filter.PageSize = 20
	}
	q := r.db.WithContext(ctx).Model(&entity.Favorite{}).Where("user_id = ?", userID)
	if filter.TargetType != "" && validTarget(filter.TargetType) {
		q = q.Where("target_type = ?", filter.TargetType)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (filter.Page - 1) * filter.PageSize
	items := make([]*entity.Favorite, 0, filter.PageSize)
	if err := q.Order("created_at DESC, id DESC").
		Offset(offset).Limit(filter.PageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *FavoriteRepo) IncrTargetCount(ctx context.Context, tx *gorm.DB, targetType entity.FavoriteTarget, targetID uint64, delta int) error {
	if targetID == 0 || delta == 0 || !validTarget(targetType) {
		return ErrInvalidFavoriteTarget
	}
	d := r.db
	if tx != nil {
		d = tx
	}
	switch targetType {
	case entity.FavoriteTargetFounder:
		if delta > 0 {
			return d.WithContext(ctx).Model(&entity.FounderProfile{}).Where("id = ?", targetID).
				UpdateColumn("favorite_count", gorm.Expr("favorite_count + ?", uint32(delta))).Error
		}
		return d.WithContext(ctx).Model(&entity.FounderProfile{}).Where("id = ?", targetID).
			UpdateColumn("favorite_count", gorm.Expr("GREATEST(favorite_count - ?, 0)", uint32(-delta))).Error
	case entity.FavoriteTargetProduct:
		if delta > 0 {
			return d.WithContext(ctx).Model(&entity.Product{}).Where("id = ?", targetID).
				UpdateColumn("favorite_count", gorm.Expr("favorite_count + ?", uint32(delta))).Error
		}
		return d.WithContext(ctx).Model(&entity.Product{}).Where("id = ?", targetID).
			UpdateColumn("favorite_count", gorm.Expr("GREATEST(favorite_count - ?, 0)", uint32(-delta))).Error
	}
	return ErrInvalidFavoriteTarget
}
