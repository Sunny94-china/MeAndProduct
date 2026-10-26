package vo

import "time"

type FounderProfileItemVO struct {
	FounderProfileVO
}

type FounderProfileDetailVO struct {
	FounderProfileVO
}

type TagVO struct {
	ID           uint64 `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Type         string `json:"type"`
	FounderCount uint32 `json:"founder_count"`
	ProductCount uint32 `json:"product_count"`
	SortOrder    int    `json:"sort_order"`
	IsActive     bool   `json:"is_active"`
}

type ProductVO struct {
	ID             uint64         `json:"id"`
	FounderID      uint64         `json:"founder_id"`
	Name           string         `json:"name"`
	Slug           string         `json:"slug"`
	Tagline        string         `json:"tagline,omitempty"`
	Description    *string        `json:"description,omitempty"`
	LogoURL        *string        `json:"logo_url,omitempty"`
	CoverURL       *string        `json:"cover_url,omitempty"`
	OfficialURL    string         `json:"official_url,omitempty"`
	GithubURL      *string        `json:"github_url,omitempty"`
	TwitterURL     *string        `json:"twitter_url,omitempty"`
	ProducthuntURL *string        `json:"producthunt_url,omitempty"`
	PricingURL     *string        `json:"pricing_url,omitempty"`
	Status         string         `json:"status"`
	IsVisible      bool           `json:"is_visible"`
	ViewCount      uint32         `json:"view_count"`
	FavoriteCount  uint32         `json:"favorite_count"`
	SortOrder      int            `json:"sort_order"`
	CustomLinks    []CustomLinkVO `json:"custom_links,omitempty"`
	Tags           []TagVO        `json:"tags,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type ProductDetailVO struct {
	ProductVO
}

type PageVO struct {
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Size  int   `json:"size"`
}

type FounderListResponse struct {
	Page PageVO                  `json:"page"`
	List []*FounderProfileItemVO `json:"list"`
}

type ProductListResponse struct {
	Page PageVO       `json:"page"`
	List []*ProductVO `json:"list"`
}

type FavoriteToggleResult struct {
	TargetType    string `json:"target_type"`
	TargetID      uint64 `json:"target_id"`
	Favorited     bool   `json:"favorited"`
	FavoriteCount uint32 `json:"favorite_count"`
}

type FavoriteItemVO struct {
	ID             uint64    `json:"id"`
	TargetType     string    `json:"target_type"`
	TargetID       uint64    `json:"target_id"`
	TargetSlug     string    `json:"target_slug,omitempty"`
	TargetTitle    string    `json:"target_title,omitempty"`
	TargetCover    *string   `json:"target_cover,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	TargetSnapshot any       `json:"target_snapshot,omitempty"`
}

type FavoriteListResponse struct {
	Page PageVO            `json:"page"`
	List []*FavoriteItemVO `json:"list"`
}

type ReportVO struct {
	ID           uint64     `json:"id"`
	ReporterID   *uint64    `json:"reporter_id,omitempty"`
	TargetType   string     `json:"target_type"`
	TargetID     uint64     `json:"target_id"`
	ReasonType   string     `json:"reason_type"`
	ReasonDetail string     `json:"reason_detail"`
	ContactInfo  string     `json:"contact_info,omitempty"`
	Status       string     `json:"status"`
	HandlerNote  *string    `json:"handler_note,omitempty"`
	HideTarget   bool       `json:"hide_target,omitempty"`
	HandlerID    *uint64    `json:"handler_id,omitempty"`
	HandledAt    *time.Time `json:"handled_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type ReportListResponse struct {
	Page PageVO      `json:"page"`
	List []*ReportVO `json:"list"`
}

type TagGroupedResponse struct {
	Groups map[string][]*TagVO `json:"groups"`
}
