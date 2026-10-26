package dto

import "strings"

// GenerateInviteReq founder 单条 / admin 批量共用字段
// *int 用于区分「字段未提交（nil，使用后端默认值）」和「用户明确传 0（非法，Validate 报错）」
type GenerateInviteReq struct {
	MaxUses     *int `json:"max_uses" binding:"omitempty"`
	ExpiresDays *int `json:"expires_days" binding:"omitempty"`
}

func (r *GenerateInviteReq) Sanitize() {}

func (r *GenerateInviteReq) Validate() map[string]string {
	detail := make(map[string]string)
	if r.MaxUses != nil {
		v := *r.MaxUses
		if v < 1 || v > 10000 {
			detail["max_uses"] = "if provided must be between 1 and 10000"
		}
	}
	if r.ExpiresDays != nil {
		v := *r.ExpiresDays
		if v < 1 || v > 3650 {
			detail["expires_days"] = "if provided must be between 1 and 3650 (days)"
		}
	}
	if len(detail) > 0 {
		return detail
	}
	return nil
}

// FounderRevokeInviteReq
type FounderRevokeInviteReq struct {
	InviteID uint64 `json:"invite_id" binding:"required,min=1"`
}

func (r *FounderRevokeInviteReq) Sanitize() {}
func (r *FounderRevokeInviteReq) Validate() map[string]string {
	detail := make(map[string]string)
	if r.InviteID == 0 {
		detail["invite_id"] = "required and must be >0"
	}
	if len(detail) > 0 {
		return detail
	}
	return nil
}

// AdminGenerateManyReq
type AdminGenerateManyReq struct {
	Count       int  `json:"count" binding:"required,min=1,max=100"`
	MaxUses     *int `json:"max_uses" binding:"omitempty"`
	ExpiresDays *int `json:"expires_days" binding:"omitempty"`
}

func (r *AdminGenerateManyReq) Sanitize() {
	if r.Count <= 0 {
		r.Count = 1
	}
}
func (r *AdminGenerateManyReq) Validate() map[string]string {
	detail := make(map[string]string)
	if r.Count < 1 || r.Count > 100 {
		detail["count"] = "must be between 1 and 100"
	}
	if r.MaxUses != nil {
		v := *r.MaxUses
		if v < 1 || v > 10000 {
			detail["max_uses"] = "if provided must be between 1 and 10000"
		}
	}
	if r.ExpiresDays != nil {
		v := *r.ExpiresDays
		if v < 1 || v > 3650 {
			detail["expires_days"] = "if provided must be between 1 and 3650 (days)"
		}
	}
	if len(detail) > 0 {
		return detail
	}
	return nil
}

// AdminRevokeManyReq
type AdminRevokeManyReq struct {
	IDs []uint64 `json:"ids" binding:"required,min=1,max=100"`
}

func (r *AdminRevokeManyReq) Sanitize() {}
func (r *AdminRevokeManyReq) Validate() map[string]string {
	detail := make(map[string]string)
	if len(r.IDs) == 0 {
		detail["ids"] = "required, at least 1 id"
	} else if len(r.IDs) > 100 {
		detail["ids"] = "at most 100 ids per request"
	} else {
		for i, id := range r.IDs {
			_ = i
			if id == 0 {
				detail["ids"] = "contains 0, which is invalid"
				break
			}
		}
	}
	if len(detail) > 0 {
		return detail
	}
	return nil
}

// ListInviteReq query 参数解析用（无需 Sanitize）
type ListInviteReq struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=500"`
	Status   string `form:"status" binding:"omitempty,oneof=active used revoked expired"`
}

func (r *ListInviteReq) Sanitize() {
	r.Status = strings.TrimSpace(strings.ToLower(r.Status))
	if r.Page < 1 {
		r.Page = 1
	}
	if r.PageSize <= 0 {
		r.PageSize = 20
	}
	if r.PageSize > 500 {
		r.PageSize = 500
	}
}
func (r *ListInviteReq) Validate() map[string]string {
	detail := make(map[string]string)
	if r.Page < 1 {
		detail["page"] = "must be >=1"
	}
	if r.PageSize < 1 || r.PageSize > 500 {
		detail["page_size"] = "must be between 1 and 500"
	}
	switch r.Status {
	case "", "active", "used", "revoked", "expired":
	default:
		detail["status"] = "must be one of active|used|revoked|expired or empty"
	}
	if len(detail) > 0 {
		return detail
	}
	return nil
}
