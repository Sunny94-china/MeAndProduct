package entity

import "time"

type LinkSource string

const (
	LinkSourceFounder LinkSource = "founder"
	LinkSourceProduct LinkSource = "product"
)

type Link struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	SourceType  LinkSource `gorm:"column:source_type;size:20;not null;index:idx_source,priority:1"`
	SourceID    uint64     `gorm:"column:source_id;not null;index:idx_source,priority:2"`
	LinkType    string     `gorm:"column:link_type;size:50;not null;index:idx_link_type"`
	URL         string     `gorm:"column:url;size:500;not null"`
	ClickCount  uint32     `gorm:"column:click_count;not null;default:0"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (Link) TableName() string { return "links" }
