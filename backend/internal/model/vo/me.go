package vo

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

// MeVO 登录用户信息（/auth/me 返回 & TokenPair 响应附带）
type MeVO struct {
	ID             uint64     `json:"id"`
	Username       string     `json:"username"`
	Email          string     `json:"email"`
	Role           string     `json:"role"`
	Status         string     `json:"status"`
	RegisterSource string     `json:"register_source"`
	AvatarURL      *string    `json:"avatar_url,omitempty"`
	LastLoginAt    *time.Time `json:"last_login_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`

	// FounderProfile（role=founder 时填充；pending/founder 均返回基础字段，is_visible=true 时返回完整公开字段）
	Founder *FounderProfileVO `json:"founder,omitempty"`
}

// FounderProfileVO 名片信息
type FounderProfileVO struct {
	ID             uint64         `json:"id"`
	UserID         uint64         `json:"user_id"`
	DisplayName    string         `json:"display_name"`
	Slug           string         `json:"slug"`
	IsVisible      bool           `json:"is_visible"`
	Title          *string        `json:"title,omitempty"`
	Bio            *string        `json:"bio,omitempty"`
	BioLong        *string        `json:"bio_long,omitempty"`
	Location       *string        `json:"location,omitempty"`
	AvatarURL      *string        `json:"avatar_url,omitempty"`
	CoverURL       *string        `json:"cover_url,omitempty"`
	ContactEmail   *string        `json:"contact_email,omitempty"`
	WebsiteURL     *string        `json:"website_url,omitempty"`
	GithubURL      *string        `json:"github_url,omitempty"`
	TwitterURL     *string        `json:"twitter_url,omitempty"`
	LinkedinURL    *string        `json:"linkedin_url,omitempty"`
	ZhihuURL       *string        `json:"zhihu_url,omitempty"`
	XiaohongshuURL *string        `json:"xiaohongshu_url,omitempty"`
	ProducthuntURL *string        `json:"producthunt_url,omitempty"`
	WechatID       *string        `json:"wechat_id,omitempty"`
	CustomLinks    []CustomLinkVO `json:"custom_links,omitempty"`
	Tags           []TagVO        `json:"tags,omitempty"`
	ViewCount      uint32         `json:"view_count"`
	FavoriteCount  uint32         `json:"favorite_count"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// CustomLinkVO 自定义外链
type CustomLinkVO struct {
	ID         uint64 `json:"id"`
	Label      string `json:"label"`
	LinkType   string `json:"link_type"`
	URL        string `json:"url"`
	ClickCount uint32 `json:"click_count"`
}

// ParseCustomLinks 解析 sql.NullString JSON 为 []CustomLinkVO
func ParseCustomLinks(ns sql.NullString) ([]CustomLinkVO, error) {
	if !ns.Valid || ns.String == "" {
		return nil, nil
	}
	var list []CustomLinkVO
	if err := json.Unmarshal([]byte(ns.String), &list); err != nil {
		return nil, errors.New("invalid custom_links json")
	}
	return list, nil
}
