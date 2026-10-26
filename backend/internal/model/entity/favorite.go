package entity

import "time"

type FavoriteTarget string

const (
	FavoriteTargetFounder FavoriteTarget = "founder"
	FavoriteTargetProduct FavoriteTarget = "product"
)

type Favorite struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement;column:id"`
	UserID     uint64         `gorm:"column:user_id;not null;uniqueIndex:uk_user_target,priority:1;index:idx_user"`
	TargetType FavoriteTarget `gorm:"column:target_type;size:20;not null;uniqueIndex:uk_user_target,priority:2;index:idx_target,priority:1"`
	TargetID   uint64         `gorm:"column:target_id;not null;uniqueIndex:uk_user_target,priority:3;index:idx_target,priority:2"`
	CreatedAt  time.Time      `gorm:"column:created_at;autoCreateTime"`
}

func (Favorite) TableName() string { return "favorites" }

var ValidFavoriteTargets = map[FavoriteTarget]bool{
	FavoriteTargetFounder: true,
	FavoriteTargetProduct: true,
}
