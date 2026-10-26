package dto

import (
	"strings"

	"meandproduct/internal/model/entity"
)

type ReportSubmitReq struct {
	TargetType   entity.ReportTarget `json:"target_type" binding:"required,oneof=founder product"`
	TargetID     uint64              `json:"target_id" binding:"required,min=1"`
	ReasonType   string              `json:"reason_type" binding:"required,oneof=spam fraud inappropriate copyright other"`
	ReasonDetail string              `json:"reason_detail" binding:"required,min=5,max=5000"`
	ContactInfo  *string             `json:"contact_info"`
}

func (r *ReportSubmitReq) Sanitize() {
	r.ReasonDetail = strings.TrimSpace(r.ReasonDetail)
	if r.ContactInfo != nil {
		v := strings.TrimSpace(*r.ContactInfo)
		r.ContactInfo = &v
	}
}

func (r *ReportSubmitReq) Validate() map[string]string {
	detail := map[string]string{}
	if r.TargetType != entity.ReportTargetFounder && r.TargetType != entity.ReportTargetProduct {
		detail["target_type"] = "must be founder or product"
	}
	if r.TargetID == 0 {
		detail["target_id"] = "required and >0"
	}
	valid := map[string]bool{"spam": true, "fraud": true, "inappropriate": true, "copyright": true, "other": true}
	if !valid[r.ReasonType] {
		detail["reason_type"] = "must be spam|fraud|inappropriate|copyright|other"
	}
	if len(r.ReasonDetail) < 5 || len(r.ReasonDetail) > 5000 {
		detail["reason_detail"] = "must be 5-5000 chars"
	}
	if r.ContactInfo != nil && *r.ContactInfo != "" && len(*r.ContactInfo) > 200 {
		detail["contact_info"] = "if provided must be <= 200"
	}
	if len(detail) > 0 {
		return detail
	}
	return nil
}

type ReportHandleReq struct {
	ReportID    uint64              `json:"report_id"`
	Status      entity.ReportStatus `json:"status" binding:"required,oneof=reviewed rejected resolved"`
	HandlerNote *string             `json:"handler_note"`
	HideTarget  bool                `json:"hide_target"`
}

func (r *ReportHandleReq) Sanitize() {
	if r.HandlerNote != nil {
		v := strings.TrimSpace(*r.HandlerNote)
		r.HandlerNote = &v
	}
}

func (r *ReportHandleReq) Validate() map[string]string {
	detail := map[string]string{}
	if r.ReportID == 0 {
		detail["report_id"] = "required and >0"
	}
	valid := map[entity.ReportStatus]bool{
		entity.ReportReviewed: true, entity.ReportRejected: true, entity.ReportResolved: true,
	}
	if !valid[r.Status] {
		detail["status"] = "must be reviewed|rejected|resolved"
	}
	if r.HandlerNote != nil && *r.HandlerNote != "" && len(*r.HandlerNote) > 5000 {
		detail["handler_note"] = "if provided must be <= 5000"
	}
	if len(detail) > 0 {
		return detail
	}
	return nil
}

type ReportListFilterReq struct {
	Status     entity.ReportStatus `json:"status"`
	ReporterID *uint64             `json:"reporter_id"`
	TargetType entity.ReportTarget `json:"target_type"`
	Page       *int                `json:"page"`
	PageSize   *int                `json:"page_size"`
}

func (r *ReportListFilterReq) Validate() map[string]string {
	detail := map[string]string{}
	if r.Status != "" {
		valid := map[entity.ReportStatus]bool{
			entity.ReportPending: true, entity.ReportReviewed: true,
			entity.ReportRejected: true, entity.ReportResolved: true,
		}
		if !valid[r.Status] {
			detail["status"] = "invalid report status"
		}
	}
	if r.TargetType != "" && r.TargetType != entity.ReportTargetFounder && r.TargetType != entity.ReportTargetProduct {
		detail["target_type"] = "invalid target_type"
	}
	if r.Page != nil && *r.Page < 1 {
		detail["page"] = "must be >= 1"
	}
	if r.PageSize != nil && (*r.PageSize < 1 || *r.PageSize > 200) {
		detail["page_size"] = "must be 1-200"
	}
	if len(detail) > 0 {
		return detail
	}
	return nil
}
