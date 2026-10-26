package dto

import "strings"

type TagCreateReq struct {
	Name      string          `json:"name" binding:"required,min=1,max=50"`
	Slug      string          `json:"slug" binding:"required,min=1,max=50"`
	Type      string          `json:"type" binding:"required,oneof=industry stage role other"`
	SortOrder *int            `json:"sort_order" binding:"omitempty"`
}

func (r *TagCreateReq) Sanitize() {
	r.Name = strings.TrimSpace(r.Name)
	r.Slug = strings.TrimSpace(r.Slug)
	r.Type = strings.TrimSpace(r.Type)
}

func (r *TagCreateReq) Validate() map[string]string {
	detail := map[string]string{}
	if r.Name == "" || len(r.Name) > 50 {
		detail["name"] = "required 1-50"
	}
	if !validTagSlug(r.Slug) {
		detail["slug"] = "only a-z0-9_- (1-50)"
	}
	validTypes := map[string]bool{"industry": true, "stage": true, "role": true, "other": true}
	if !validTypes[r.Type] {
		detail["type"] = "must be industry|stage|role|other"
	}
	if r.SortOrder != nil {
		if *r.SortOrder < -10000 || *r.SortOrder > 10000 {
			detail["sort_order"] = "if provided must be -10000..10000"
		}
	}
	if len(detail) > 0 {
		return detail
	}
	return nil
}

type TagUpdateReq struct {
	Name      *string `json:"name"`
	Slug      *string `json:"slug"`
	Type      *string `json:"type"`
	SortOrder *int    `json:"sort_order"`
	IsActive  *bool   `json:"is_active"`
}

func (r *TagUpdateReq) Sanitize() {
	if r.Name != nil {
		v := strings.TrimSpace(*r.Name)
		r.Name = &v
	}
	if r.Slug != nil {
		v := strings.TrimSpace(*r.Slug)
		r.Slug = &v
	}
	if r.Type != nil {
		v := strings.TrimSpace(*r.Type)
		r.Type = &v
	}
}

func (r *TagUpdateReq) Validate() map[string]string {
	detail := map[string]string{}
	if r.Name != nil {
		if *r.Name == "" || len(*r.Name) > 50 {
			detail["name"] = "if provided 1-50"
		}
	}
	if r.Slug != nil && !validTagSlug(*r.Slug) {
		detail["slug"] = "if provided 1-50 a-z0-9_-"
	}
	if r.Type != nil {
		validTypes := map[string]bool{"industry": true, "stage": true, "role": true, "other": true}
		if !validTypes[*r.Type] {
			detail["type"] = "if provided must be industry|stage|role|other"
		}
	}
	if r.SortOrder != nil && (*r.SortOrder < -10000 || *r.SortOrder > 10000) {
		detail["sort_order"] = "if provided must be -10000..10000"
	}
	if len(detail) > 0 {
		return detail
	}
	return nil
}

func validTagSlug(s string) bool {
	if len(s) < 1 || len(s) > 50 {
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
