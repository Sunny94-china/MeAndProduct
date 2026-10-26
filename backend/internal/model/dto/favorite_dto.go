package dto

import (
	"meandproduct/internal/model/entity"
)

type FavoriteToggleReq struct {
	TargetType entity.FavoriteTarget `json:"target_type" binding:"required,oneof=founder product"`
	TargetID   uint64                `json:"target_id" binding:"required,min=1"`
}

func (r *FavoriteToggleReq) Sanitize() {}

func (r *FavoriteToggleReq) Validate() map[string]string {
	detail := map[string]string{}
	if !entity.ValidFavoriteTargets[r.TargetType] {
		detail["target_type"] = "must be founder or product"
	}
	if r.TargetID == 0 {
		detail["target_id"] = "required and >0"
	}
	if len(detail) > 0 {
		return detail
	}
	return nil
}

type FavoriteListReq struct {
	TargetType entity.FavoriteTarget `json:"target_type" binding:"omitempty,oneof=founder product"`
	Page       *int                  `json:"page"`
	PageSize   *int                  `json:"page_size"`
}

func (r *FavoriteListReq) Sanitize() {}

func (r *FavoriteListReq) Validate() map[string]string {
	detail := map[string]string{}
	if r.TargetType != "" && !entity.ValidFavoriteTargets[r.TargetType] {
		detail["target_type"] = "if provided must be founder|product"
	}
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
