package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"meandproduct/internal/middleware"
	"meandproduct/internal/model/dto"
	"meandproduct/internal/model/entity"
	"meandproduct/internal/pkg/response"
	"meandproduct/internal/service"
)

// InviteHandler 邀请码管理 HTTP
type InviteHandler struct {
	svc *service.InviteService
}

func NewInviteHandler(svc *service.InviteService) *InviteHandler {
	return &InviteHandler{svc: svc}
}

// inviteErrMap 把 invite service Err* 映射 HTTP 码
func inviteErrHTTP(err error) (int, string) {
	switch {
	case err == nil:
		return 0, ""
	case errors.Is(err, service.ErrInviteCodeNotFound):
		return http.StatusNotFound, "invite code not found"
	case errors.Is(err, service.ErrInviteCodeNotYours):
		return http.StatusForbidden, "invite code does not belong to you"
	case errors.Is(err, service.ErrInviteCodeAlreadyUsed):
		return http.StatusUnprocessableEntity, "invite code already used, cannot revoke"
	case errors.Is(err, service.ErrInviteInvalidMaxUses):
		return http.StatusUnprocessableEntity, "invalid max_uses"
	case errors.Is(err, service.ErrInviteInvalidExpiresAt):
		return http.StatusUnprocessableEntity, "invalid expires_at (should be in future)"
	case errors.Is(err, service.ErrInviteInvalidCount):
		return http.StatusUnprocessableEntity, "invalid count, must be 1~100"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

// ================================================================
// Founder Routes
// ================================================================

// FounderGenerate POST /api/v1/founder/invites
func (h *InviteHandler) FounderGenerate(c *gin.Context) {
	var req dto.GenerateInviteReq
	if ok := BindValidate(c, &req); !ok {
		return
	}
	cu := middleware.GetCurrentUser(c)
	if cu == nil {
		response.Fail(c, http.StatusUnauthorized, "login required")
		return
	}
	var maxUses int
	if req.MaxUses != nil {
		maxUses = *req.MaxUses
	}
	var expiresAt time.Time
	if req.ExpiresDays != nil {
		expiresAt = time.Now().Add(time.Duration(*req.ExpiresDays) * 24 * time.Hour)
	}

	ic, err := h.svc.FounderGenerate(c.Request.Context(), cu.UserID, maxUses, expiresAt)
	if code, msg := inviteErrHTTP(err); code != 0 {
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, gin.H{"invite": inviteToVO(ic)})
}

// FounderList GET /api/v1/founder/invites
func (h *InviteHandler) FounderList(c *gin.Context) {
	var req dto.ListInviteReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	req.Sanitize()
	if detail := req.Validate(); detail != nil {
		response.FailWithDetail(c, http.StatusUnprocessableEntity, "validation failed", detail)
		return
	}
	cu := middleware.GetCurrentUser(c)
	if cu == nil {
		response.Fail(c, http.StatusUnauthorized, "login required")
		return
	}
	list, total, err := h.svc.FounderList(c.Request.Context(), cu.UserID, req.Page, req.PageSize)
	if code, msg := inviteErrHTTP(err); code != 0 {
		response.Fail(c, code, msg)
		return
	}
	vos := make([]any, 0, len(list))
	for _, ic := range list {
		vos = append(vos, inviteToVO(ic))
	}
	response.OK(c, gin.H{"items": vos, "total": total, "page": req.Page, "page_size": req.PageSize})
}

// FounderRevoke POST /api/v1/founder/invites/revoke
func (h *InviteHandler) FounderRevoke(c *gin.Context) {
	var req dto.FounderRevokeInviteReq
	if ok := BindValidate(c, &req); !ok {
		return
	}
	cu := middleware.GetCurrentUser(c)
	if cu == nil {
		response.Fail(c, http.StatusUnauthorized, "login required")
		return
	}
	err := h.svc.FounderRevoke(c.Request.Context(), cu.UserID, req.InviteID)
	if code, msg := inviteErrHTTP(err); code != 0 {
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, nil)
}

// ================================================================
// Admin Routes
// ================================================================

// AdminList GET /api/v1/admin/invites
func (h *InviteHandler) AdminList(c *gin.Context) {
	var req dto.ListInviteReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	req.Sanitize()
	if detail := req.Validate(); detail != nil {
		response.FailWithDetail(c, http.StatusUnprocessableEntity, "validation failed", detail)
		return
	}
	var status entity.InviteCodeStatus
	if req.Status != "" {
		status = entity.InviteCodeStatus(req.Status)
	}
	list, total, err := h.svc.AdminList(c.Request.Context(), status, req.Page, req.PageSize)
	if code, msg := inviteErrHTTP(err); code != 0 {
		response.Fail(c, code, msg)
		return
	}
	vos := make([]any, 0, len(list))
	for _, ic := range list {
		vos = append(vos, inviteToVO(ic))
	}
	response.OK(c, gin.H{"items": vos, "total": total, "page": req.Page, "page_size": req.PageSize})
}

// AdminGenerateMany POST /api/v1/admin/invites/generate
func (h *InviteHandler) AdminGenerateMany(c *gin.Context) {
	var req dto.AdminGenerateManyReq
	if ok := BindValidate(c, &req); !ok {
		return
	}
	cu := middleware.GetCurrentUser(c)
	if cu == nil {
		response.Fail(c, http.StatusUnauthorized, "login required")
		return
	}
	var maxUses int
	if req.MaxUses != nil {
		maxUses = *req.MaxUses
	}
	var expiresAt time.Time
	if req.ExpiresDays != nil {
		expiresAt = time.Now().Add(time.Duration(*req.ExpiresDays) * 24 * time.Hour)
	}
	list, err := h.svc.AdminGenerateMany(c.Request.Context(), cu.UserID, req.Count, maxUses, expiresAt)
	if code, msg := inviteErrHTTP(err); code != 0 {
		response.Fail(c, code, msg)
		return
	}
	vos := make([]any, 0, len(list))
	for _, ic := range list {
		vos = append(vos, inviteToVO(ic))
	}
	response.OK(c, gin.H{"items": vos})
}

// AdminRevokeMany POST /api/v1/admin/invites/revoke
func (h *InviteHandler) AdminRevokeMany(c *gin.Context) {
	var req dto.AdminRevokeManyReq
	if ok := BindValidate(c, &req); !ok {
		return
	}
	rows, err := h.svc.AdminRevokeMany(c.Request.Context(), req.IDs)
	if code, msg := inviteErrHTTP(err); code != 0 {
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, gin.H{"revoked": rows})
}

// inviteToVO 精简输出；避免暴露不必要的内部字段
func inviteToVO(ic *entity.InviteCode) map[string]any {
	m := map[string]any{
		"id":         ic.ID,
		"code":       ic.Code,
		"created_by": ic.CreatedBy,
		"max_uses":   ic.MaxUses,
		"used_count": ic.UsedCount,
		"status":     ic.Status,
		"created_at": ic.CreatedAt,
	}
	if ic.ExpiresAt != nil {
		m["expires_at"] = *ic.ExpiresAt
	}
	if ic.UsedAt != nil {
		m["used_at"] = *ic.UsedAt
	}
	return m
}
