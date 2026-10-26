package dto

import (
	"strconv"
	"strings"

	"meandproduct/internal/model/entity"
)

type LinkRow struct {
	LinkType string `json:"link_type"`
	URL      string `json:"url"`
}

type ProductCreateReq struct {
	Name           string              `json:"name" binding:"required,min=1,max=200"`
	Slug           string              `json:"slug" binding:"required,min=3,max=200"`
	Tagline        string              `json:"tagline"`
	Description    *string             `json:"description"`
	LogoURL        *string             `json:"logo_url"`
	CoverURL       *string             `json:"cover_url"`
	OfficialURL    string              `json:"official_url"`
	GithubURL      *string             `json:"github_url"`
	TwitterURL     *string             `json:"twitter_url"`
	ProducthuntURL *string             `json:"producthunt_url"`
	PricingURL     *string             `json:"pricing_url"`
	Status         entity.ProductStage `json:"status"`
	IsVisible      *bool               `json:"is_visible"`
	CustomLinks    []LinkRow           `json:"custom_links"`
	TagIDs         []uint64            `json:"tag_ids"`
}

func (r *ProductCreateReq) Sanitize() {
	r.Name = strings.TrimSpace(r.Name)
	r.Slug = strings.TrimSpace(r.Slug)
	r.Tagline = strings.TrimSpace(r.Tagline)
	if r.Description != nil {
		v := strings.TrimSpace(*r.Description)
		r.Description = &v
	}
	if r.LogoURL != nil {
		v := strings.TrimSpace(*r.LogoURL)
		r.LogoURL = &v
	}
	if r.CoverURL != nil {
		v := strings.TrimSpace(*r.CoverURL)
		r.CoverURL = &v
	}
	r.OfficialURL = strings.TrimSpace(r.OfficialURL)
	if r.GithubURL != nil {
		v := strings.TrimSpace(*r.GithubURL)
		r.GithubURL = &v
	}
	if r.TwitterURL != nil {
		v := strings.TrimSpace(*r.TwitterURL)
		r.TwitterURL = &v
	}
	if r.ProducthuntURL != nil {
		v := strings.TrimSpace(*r.ProducthuntURL)
		r.ProducthuntURL = &v
	}
	if r.PricingURL != nil {
		v := strings.TrimSpace(*r.PricingURL)
		r.PricingURL = &v
	}
	for i := range r.CustomLinks {
		r.CustomLinks[i].LinkType = strings.TrimSpace(r.CustomLinks[i].LinkType)
		r.CustomLinks[i].URL = strings.TrimSpace(r.CustomLinks[i].URL)
	}
	if r.Status == "" {
		r.Status = entity.StageIdea
	}
}

func (r *ProductCreateReq) Validate() map[string]string {
	detail := map[string]string{}
	if r.Name == "" || len(r.Name) > 200 {
		detail["name"] = "required 1-200"
	}
	if !validProductSlug(r.Slug) {
		detail["slug"] = "only a-z0-9_- (3-200 chars)"
	}
	validStages := map[entity.ProductStage]bool{
		entity.StageIdea: true, entity.StageBeta: true, entity.StageLaunched: true,
		entity.StageGrowing: true, entity.StageArchived: true,
	}
	if r.Status != "" && !validStages[r.Status] {
		detail["status"] = "invalid stage"
	}
	for i, link := range r.CustomLinks {
		if link.LinkType == "" || len(link.LinkType) > 50 {
			detail["custom_links["+itoa(i)+"].link_type"] = "required and <= 50"
		}
		if link.URL == "" || len(link.URL) > 500 {
			detail["custom_links["+itoa(i)+"].url"] = "required and <= 500"
		}
	}
	if len(detail) > 0 {
		return detail
	}
	return nil
}

type ProductUpdateReq struct {
	Name           *string              `json:"name"`
	Slug           *string              `json:"slug"`
	Tagline        *string              `json:"tagline"`
	Description    *string              `json:"description"`
	LogoURL        *string              `json:"logo_url"`
	CoverURL       *string              `json:"cover_url"`
	OfficialURL    *string              `json:"official_url"`
	GithubURL      *string              `json:"github_url"`
	TwitterURL     *string              `json:"twitter_url"`
	ProducthuntURL *string              `json:"producthunt_url"`
	PricingURL     *string              `json:"pricing_url"`
	Status         *entity.ProductStage `json:"status"`
	IsVisible      *bool                `json:"is_visible"`
	CustomLinks    *[]LinkRow           `json:"custom_links"`
	TagIDs         *[]uint64            `json:"tag_ids"`
}

func (r *ProductUpdateReq) Sanitize() {
	if r.Name != nil {
		v := strings.TrimSpace(*r.Name)
		r.Name = &v
	}
	if r.Slug != nil {
		v := strings.TrimSpace(*r.Slug)
		r.Slug = &v
	}
	if r.Tagline != nil {
		v := strings.TrimSpace(*r.Tagline)
		r.Tagline = &v
	}
	if r.Description != nil {
		v := strings.TrimSpace(*r.Description)
		r.Description = &v
	}
	if r.LogoURL != nil {
		v := strings.TrimSpace(*r.LogoURL)
		r.LogoURL = &v
	}
	if r.CoverURL != nil {
		v := strings.TrimSpace(*r.CoverURL)
		r.CoverURL = &v
	}
	if r.OfficialURL != nil {
		v := strings.TrimSpace(*r.OfficialURL)
		r.OfficialURL = &v
	}
	if r.GithubURL != nil {
		v := strings.TrimSpace(*r.GithubURL)
		r.GithubURL = &v
	}
	if r.TwitterURL != nil {
		v := strings.TrimSpace(*r.TwitterURL)
		r.TwitterURL = &v
	}
	if r.ProducthuntURL != nil {
		v := strings.TrimSpace(*r.ProducthuntURL)
		r.ProducthuntURL = &v
	}
	if r.PricingURL != nil {
		v := strings.TrimSpace(*r.PricingURL)
		r.PricingURL = &v
	}
	if r.CustomLinks != nil {
		for i := range *r.CustomLinks {
			(*r.CustomLinks)[i].LinkType = strings.TrimSpace((*r.CustomLinks)[i].LinkType)
			(*r.CustomLinks)[i].URL = strings.TrimSpace((*r.CustomLinks)[i].URL)
		}
	}
}

func (r *ProductUpdateReq) Validate() map[string]string {
	detail := map[string]string{}
	if r.Name != nil {
		if *r.Name == "" || len(*r.Name) > 200 {
			detail["name"] = "if provided must be 1-200"
		}
	}
	if r.Slug != nil && !validProductSlug(*r.Slug) {
		detail["slug"] = "if provided must be 3-200 chars a-z0-9_-"
	}
	validStages := map[entity.ProductStage]bool{
		entity.StageIdea: true, entity.StageBeta: true, entity.StageLaunched: true,
		entity.StageGrowing: true, entity.StageArchived: true,
	}
	if r.Status != nil && *r.Status != "" && !validStages[*r.Status] {
		detail["status"] = "invalid stage"
	}
	if r.CustomLinks != nil {
		for i, link := range *r.CustomLinks {
			if link.LinkType == "" || len(link.LinkType) > 50 {
				detail["custom_links["+itoa(i)+"].link_type"] = "if provided required and <= 50"
			}
			if link.URL == "" || len(link.URL) > 500 {
				detail["custom_links["+itoa(i)+"].url"] = "if provided required and <= 500"
			}
		}
	}
	if len(detail) > 0 {
		return detail
	}
	return nil
}

type ProductListFilterReq struct {
	TagIDs   []uint64            `json:"tag_ids"`
	Stage    entity.ProductStage `json:"stage"`
	Page     *int                `json:"page"`
	PageSize *int                `json:"page_size"`
	SortBy   string              `json:"sort_by" binding:"omitempty,oneof=favorite_count view_count created_at"`
}

func (r *ProductListFilterReq) Sanitize() {
	if r.SortBy == "" {
		r.SortBy = "created_at"
	}
}

func (r *ProductListFilterReq) Validate() map[string]string {
	detail := map[string]string{}
	if r.Page != nil && *r.Page < 1 {
		detail["page"] = "must be >= 1"
	}
	if r.PageSize != nil && (*r.PageSize < 1 || *r.PageSize > 100) {
		detail["page_size"] = "must be 1-100"
	}
	if len(detail) > 0 {
		return detail
	}
	return nil
}

type ProductReorderReq struct {
	OrderedIDs []uint64 `json:"ordered_ids" binding:"required,max=500"`
}

func (r *ProductReorderReq) Sanitize() {
	seen := map[uint64]bool{}
	out := make([]uint64, 0, len(r.OrderedIDs))
	for _, id := range r.OrderedIDs {
		if id == 0 || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	r.OrderedIDs = out
}

func (r *ProductReorderReq) Validate() map[string]string {
	detail := map[string]string{}
	if len(r.OrderedIDs) == 0 {
		detail["ordered_ids"] = "required and not empty"
	}
	if len(r.OrderedIDs) > 500 {
		detail["ordered_ids"] = "too many items (max 500)"
	}
	if len(detail) > 0 {
		return detail
	}
	return nil
}

func validProductSlug(s string) bool {
	if len(s) < 3 || len(s) > 200 {
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

var _ = entity.Product{}

func itoa(i int) string { return strconv.Itoa(i) }
