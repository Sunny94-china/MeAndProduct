package entity

import "time"

type TagType string

const (
	TagTypeIndustry TagType = "industry"
	TagTypeStage    TagType = "stage"
	TagTypeRole     TagType = "role"
	TagTypeOther    TagType = "other"
)

type Tag struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	Name          string     `gorm:"column:name;size:50;not null;uniqueIndex:uk_name"`
	Slug          string     `gorm:"column:slug;size:50;not null;uniqueIndex:uk_slug"`
	Type          TagType    `gorm:"column:type;size:20;not null;default:industry;index:idx_type_active,priority:1"`
	IconURL       *string    `gorm:"column:icon_url;size:500"`
	SortOrder     int        `gorm:"column:sort_order;not null;default:0"`
	FounderCount  uint32     `gorm:"column:founder_count;not null;default:0"`
	ProductCount  uint32     `gorm:"column:product_count;not null;default:0"`
	IsActive      bool       `gorm:"column:is_active;not null;default:true;index:idx_type_active,priority:2"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	Founders      []*FounderProfile `gorm:"many2many:founder_tags;joinForeignKey:TagID;joinReferences:FounderID"`
	Products      []*Product `gorm:"many2many:product_tags;joinForeignKey:TagID;joinReferences:ProductID"`
}

func (Tag) TableName() string { return "tags" }

type FounderTag struct {
	FounderID uint64    `gorm:"primaryKey;column:founder_id"`
	TagID     uint64    `gorm:"primaryKey;column:tag_id;index:idx_tag"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (FounderTag) TableName() string { return "founder_tags" }

type ProductTag struct {
	ProductID uint64    `gorm:"primaryKey;column:product_id"`
	TagID     uint64    `gorm:"primaryKey;column:tag_id;index:idx_tag"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (ProductTag) TableName() string { return "product_tags" }
