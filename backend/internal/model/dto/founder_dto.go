package dto

import (
	"strconv"
	"strings"

	"meandproduct/internal/model/entity"
)

type FounderUpdateProfileReq struct {
	DisplayName    *string   `json:"display_name"`
	Title          *string   `json:"title"`
	Bio            *string   `json:"bio"`
	Location       *string   `json:"location"`
	ContactEmail   *string   `json:"contact_email"`
	AvatarURL      *string   `json:"avatar_url"`
	CoverURL       *string   `json:"cover_url"`
	ProducthuntURL *string   `json:"producthunt_url"`
	WechatID       *string   `json:"wechat_id"`
	BioLong        *string   `json:"bio_long"`
	WebsiteURL     *string   `json:"website_url"`
	Slug           *string   `json:"slug"`
	TagIDs         *[]uint64 `json:"tag_ids"`
	CustomLinks    []LinkRow `json:"custom_links"`
	IsVisible      *bool     `json:"is_visible"`
}

func (r *FounderUpdateProfileReq) Sanitize() {
	trimSP := func(s *string) *string {
		if s == nil {
			return nil
		}
		v := strings.TrimSpace(*s)
		return &v
	}
	r.DisplayName = trimSP(r.DisplayName)
	r.Title = trimSP(r.Title)
	r.Bio = trimSP(r.Bio)
	r.Location = trimSP(r.Location)
	r.ContactEmail = trimSP(r.ContactEmail)
	r.AvatarURL = trimSP(r.AvatarURL)
	r.CoverURL = trimSP(r.CoverURL)
	r.ProducthuntURL = trimSP(r.ProducthuntURL)
	r.WechatID = trimSP(r.WechatID)
	r.BioLong = trimSP(r.BioLong)
	r.WebsiteURL = trimSP(r.WebsiteURL)
	r.Slug = trimSP(r.Slug)
	for i := range r.CustomLinks {
		r.CustomLinks[i].LinkType = strings.TrimSpace(r.CustomLinks[i].LinkType)
		r.CustomLinks[i].URL = strings.TrimSpace(r.CustomLinks[i].URL)
	}
}

func (r *FounderUpdateProfileReq) Validate() map[string]string {
	detail := map[string]string{}
	if r.DisplayName != nil {
		if *r.DisplayName == "" || len(*r.DisplayName) > 100 {
			detail["display_name"] = "if provided must be 1-100 chars"
		}
	}
	if r.Title != nil && len(*r.Title) > 200 {
		detail["title"] = "if provided must be <= 200 chars"
	}
	if r.Bio != nil && len(*r.Bio) > 5000 {
		detail["bio"] = "if provided must be <= 5000 chars"
	}
	if r.Location != nil && len(*r.Location) > 200 {
		detail["location"] = "if provided must be <= 200 chars"
	}
	if r.ContactEmail != nil && *r.ContactEmail != "" {
		if len(*r.ContactEmail) > 200 || !strings.Contains(*r.ContactEmail, "@") {
			detail["contact_email"] = "invalid email"
		}
	}
	if r.AvatarURL != nil && *r.AvatarURL != "" && len(*r.AvatarURL) > 1000 {
		detail["avatar_url"] = "if provided must be <= 1000 chars"
	}
	if r.CoverURL != nil && *r.CoverURL != "" && len(*r.CoverURL) > 1000 {
		detail["cover_url"] = "if provided must be <= 1000 chars"
	}
	if r.ProducthuntURL != nil && *r.ProducthuntURL != "" && len(*r.ProducthuntURL) > 1000 {
		detail["producthunt_url"] = "if provided must be <= 1000 chars"
	}
	if r.WechatID != nil && *r.WechatID != "" && len(*r.WechatID) > 100 {
		detail["wechat_id"] = "if provided must be <= 100 chars"
	}
	if r.BioLong != nil && len(*r.BioLong) > 20000 {
		detail["bio_long"] = "if provided must be <= 20000 chars"
	}
	if r.WebsiteURL != nil && *r.WebsiteURL != "" && len(*r.WebsiteURL) > 500 {
		detail["website_url"] = "if provided must be <= 500 chars"
	}
	if r.Slug != nil && *r.Slug != "" && !validSlug(*r.Slug) {
		detail["slug"] = "only a-z0-9_- (3-50 chars)"
	}
	for i, link := range r.CustomLinks {
		if link.LinkType == "" {
			detail["custom_links["+intStr(i)+"].link_type"] = "required"
		} else if len(link.LinkType) > 50 {
			detail["custom_links["+intStr(i)+"].link_type"] = "must be <= 50"
		}
		if link.URL == "" {
			detail["custom_links["+intStr(i)+"].url"] = "required"
		} else if len(link.URL) > 500 {
			detail["custom_links["+intStr(i)+"].url"] = "must be <= 500"
		}
	}
	if len(detail) > 0 {
		return detail
	}
	return nil
}

type FounderListFilterReq struct {
	TagIDs   []uint64 `json:"tag_ids"`
	Page     *int     `json:"page" binding:"omitempty"`
	PageSize *int     `json:"page_size" binding:"omitempty"`
	SortBy   string   `json:"sort_by" binding:"omitempty,oneof=favorite_count view_count created_at"`
}

func (r *FounderListFilterReq) Sanitize() {
	if r.SortBy == "" {
		r.SortBy = "created_at"
	}
}

func (r *FounderListFilterReq) Validate() map[string]string {
	detail := map[string]string{}
	if r.Page != nil && *r.Page < 1 {
		detail["page"] = "if provided must be >=1"
	}
	if r.PageSize != nil && (*r.PageSize < 1 || *r.PageSize > 100) {
		detail["page_size"] = "if provided must be 1-100"
	}
	if len(detail) > 0 {
		return detail
	}
	return nil
}

type FounderApplyProfileReq struct {
	DisplayName    string    `json:"display_name" binding:"required,min=1,max=100"`
	Title          string    `json:"title"`
	Bio            string    `json:"bio"`
	Slug           string    `json:"slug" binding:"required,min=3,max=50"`
	Location       string    `json:"location"`
	ContactEmail   string    `json:"contact_email"`
	AvatarURL      string    `json:"avatar_url"`
	CoverURL       string    `json:"cover_url"`
	ProducthuntURL string    `json:"producthunt_url"`
	WechatID       string    `json:"wechat_id"`
	BioLong        string    `json:"bio_long"`
	WebsiteURL     string    `json:"website_url"`
	TagIDs         []uint64  `json:"tag_ids"`
	CustomLinks    []LinkRow `json:"custom_links"`
}

func (r *FounderApplyProfileReq) Sanitize() {
	r.DisplayName = strings.TrimSpace(r.DisplayName)
	r.Title = strings.TrimSpace(r.Title)
	r.Bio = strings.TrimSpace(r.Bio)
	r.Slug = strings.TrimSpace(r.Slug)
	r.Location = strings.TrimSpace(r.Location)
	r.ContactEmail = strings.TrimSpace(r.ContactEmail)
	r.AvatarURL = strings.TrimSpace(r.AvatarURL)
	r.CoverURL = strings.TrimSpace(r.CoverURL)
	r.ProducthuntURL = strings.TrimSpace(r.ProducthuntURL)
	r.WechatID = strings.TrimSpace(r.WechatID)
	r.BioLong = strings.TrimSpace(r.BioLong)
	r.WebsiteURL = strings.TrimSpace(r.WebsiteURL)
	for i := range r.CustomLinks {
		r.CustomLinks[i].LinkType = strings.TrimSpace(r.CustomLinks[i].LinkType)
		r.CustomLinks[i].URL = strings.TrimSpace(r.CustomLinks[i].URL)
	}
}

func (r *FounderApplyProfileReq) Validate() map[string]string {
	detail := map[string]string{}
	if r.DisplayName == "" || len(r.DisplayName) > 100 {
		detail["display_name"] = "required 1-100"
	}
	if !validSlug(r.Slug) {
		detail["slug"] = "only a-z0-9_- (3-50 chars)"
	}
	if r.ContactEmail != "" && !strings.Contains(r.ContactEmail, "@") {
		detail["contact_email"] = "invalid email"
	}
	if len(r.BioLong) > 20000 {
		detail["bio_long"] = "must be <= 20000 chars"
	}
	if len(r.WebsiteURL) > 500 {
		detail["website_url"] = "must be <= 500 chars"
	}
	for i, link := range r.CustomLinks {
		if link.LinkType == "" || len(link.LinkType) > 50 {
			detail["custom_links["+intStr(i)+"].link_type"] = "required and <= 50"
		}
		if link.URL == "" || len(link.URL) > 500 {
			detail["custom_links["+intStr(i)+"].url"] = "required and <= 500"
		}
	}
	if len(detail) > 0 {
		return detail
	}
	return nil
}

func validSlug(s string) bool {
	if len(s) < 3 || len(s) > 50 {
		return false
	}
	for _, c := range s {
		ok := (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_'
		if !ok {
			return false
		}
	}
	return true
}

func intStr(i int) string {
	return strconv.Itoa(i)
}

func strconvItoa(i int) string {
	return strconv.Itoa(i)
}

var _ = entity.FounderProfile{}
