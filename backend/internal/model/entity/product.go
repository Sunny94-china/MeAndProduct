package entity

import (
	"database/sql"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type ProductStage string

const (
	StageIdea     ProductStage = "idea"
	StageBeta     ProductStage = "beta"
	StageLaunched ProductStage = "launched"
	StageGrowing  ProductStage = "growing"
	StageArchived ProductStage = "archived"
)

type Product struct {
	ID             uint64         `gorm:"primaryKey;autoIncrement;column:id"`
	FounderID      uint64         `gorm:"column:founder_id;not null;index:idx_founder_id;uniqueIndex:uk_founder_slug,priority:1"`
	Name           string         `gorm:"column:name;size:200;not null"`
	Slug           string         `gorm:"column:slug;size:200;not null;uniqueIndex:uk_founder_slug,priority:2"`
	LogoURL        *string        `gorm:"column:logo_url;size:500"`
	CoverURL       *string        `gorm:"column:cover_url;size:500"`
	Tagline        string         `gorm:"column:tagline;size:200;not null"`
	Description    *string        `gorm:"column:description;type:text"`
	OfficialURL    string         `gorm:"column:official_url;size:500;not null"`
	GithubURL      *string        `gorm:"column:github_url;size:500"`
	ProducthuntURL *string        `gorm:"column:producthunt_url;size:500"`
	TwitterURL     *string        `gorm:"column:twitter_url;size:500"`
	Status         ProductStage   `gorm:"column:status;size:20;not null;default:idea;index:idx_status"`
	LaunchDate     *time.Time     `gorm:"column:launch_date;default:null"`
	PricingURL     *string        `gorm:"column:pricing_url;size:500"`
	CustomLinks    sql.NullString `gorm:"column:custom_links;type:json"`
	ViewCount      uint32         `gorm:"column:view_count;not null;default:0;index:idx_view_count,sort:desc"`
	FavoriteCount  uint32         `gorm:"column:favorite_count;not null;default:0"`
	SortOrder      int            `gorm:"column:sort_order;not null;default:0"`
	IsVisible      bool           `gorm:"column:is_visible;not null;default:true"`
	CreatedAt      time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index;default:null"`
	Tags           []*Tag         `gorm:"many2many:product_tags;joinForeignKey:ProductID;joinReferences:TagID"`
}

func (Product) TableName() string { return "products" }

func (p *Product) SetCustomLinks(links any) error {
	if links == nil {
		p.CustomLinks = sql.NullString{Valid: false}
		return nil
	}
	b, err := json.Marshal(links)
	if err != nil {
		return err
	}
	p.CustomLinks = sql.NullString{String: string(b), Valid: true}
	return nil
}

func (p *Product) ParseCustomLinks(dst any) error {
	if !p.CustomLinks.Valid {
		return nil
	}
	return json.Unmarshal([]byte(p.CustomLinks.String), dst)
}
