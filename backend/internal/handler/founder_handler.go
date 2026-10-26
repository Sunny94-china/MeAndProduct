package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"meandproduct/internal/middleware"
	"meandproduct/internal/model/dto"
	"meandproduct/internal/model/vo"
	"meandproduct/internal/pkg/response"
	"meandproduct/internal/service"
)

type FounderHandler struct {
	svc *service.FounderProfileService
}

func NewFounderHandler(s *service.FounderProfileService) *FounderHandler {
	return &FounderHandler{svc: s}
}

func contentErrHTTP(err error) (int, string) {
	switch {
	case err == nil:
		return 0, ""
	case errors.Is(err, service.ErrFounderNotFound):
		return http.StatusNotFound, "founder not found"
	case errors.Is(err, service.ErrProductNotFound):
		return http.StatusNotFound, "product not found"
	case errors.Is(err, service.ErrTagNotFound):
		return http.StatusNotFound, "tag not found"
	case errors.Is(err, service.ErrReportNotFound):
		return http.StatusNotFound, "report not found"
	case errors.Is(err, service.ErrInvalidTarget):
		return http.StatusBadRequest, "invalid target type or id"
	case errors.Is(err, service.ErrInvalidReportStatus):
		return http.StatusUnprocessableEntity, "invalid report status transition"
	case errors.Is(err, service.ErrFounderSlugConflict):
		return http.StatusConflict, "founder slug already in use"
	case errors.Is(err, service.ErrTagSlugConflict):
		return http.StatusConflict, "tag slug or name already in use"
	case errors.Is(err, service.ErrConflictUsernameEmail):
		return http.StatusConflict, "username or email already exists"
	case errors.Is(err, service.ErrInvalidCredentials):
		return http.StatusUnauthorized, "invalid username/email or password"
	case errors.Is(err, service.ErrUserSuspendedOrPending):
		return http.StatusForbidden, "account suspended or pending, please contact admin"
	case errors.Is(err, service.ErrInviteCodeInvalid):
		return http.StatusUnprocessableEntity, "invalid or unavailable invite code"
	case errors.Is(err, service.ErrRefreshTokenInvalid):
		return http.StatusUnauthorized, "invalid or revoked refresh token"
	case errors.Is(err, service.ErrUserNotFound):
		return http.StatusNotFound, "user not found"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

func failContent(c *gin.Context, err error) {
	code, msg := contentErrHTTP(err)
	if code == http.StatusInternalServerError {
		c.Error(err)
	}
	response.Fail(c, code, msg)
}

func (h *FounderHandler) PublicList(c *gin.Context) {
	var req dto.FounderListFilterReq
	if !BindValidate(c, &req) {
		return
	}
	if req.Page == nil {
		v := 1
		req.Page = &v
	}
	if req.PageSize == nil {
		v := 30
		req.PageSize = &v
	}
	if *req.Page < 1 {
		v := 1
		req.Page = &v
	}
	if *req.PageSize < 1 || *req.PageSize > 100 {
		v := 30
		req.PageSize = &v
	}
	list, total, err := h.svc.List(c.Request.Context(), req)
	if err != nil {
		failContent(c, err)
		return
	}
	response.OK(c, vo.FounderListResponse{
		Page: vo.PageVO{Total: total, Page: *req.Page, Size: *req.PageSize},
		List: list,
	})
}

func (h *FounderHandler) PublicGetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	fp, err := h.svc.GetBySlug(c.Request.Context(), slug, false, nil)
	if err != nil {
		failContent(c, err)
		return
	}
	response.OK(c, fp)
}

// ====== FOUNDER / ADMIN ======

func (h *FounderHandler) GetMe(c *gin.Context) {
	userID := mustCurrentUserID(c)
	fp, err := h.svc.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		failContent(c, err)
		return
	}
	response.OK(c, fp)
}

func mustCurrentUserID(c *gin.Context) uint64 {
	u := middleware.GetCurrentUser(c)
	if u == nil {
		return 0
	}
	return u.UserID
}

func mustCurrentUserRole(c *gin.Context) string {
	u := middleware.GetCurrentUser(c)
	if u == nil {
		return ""
	}
	return u.Role
}

func (h *FounderHandler) UpdateMe(c *gin.Context) {
	userID := mustCurrentUserID(c)
	role := mustCurrentUserRole(c)
	var req dto.FounderUpdateProfileReq
	if !BindValidate(c, &req) {
		return
	}
	fp, err := h.svc.Update(c.Request.Context(), userID, role, &req)
	if err != nil {
		failContent(c, err)
		return
	}
	response.OK(c, fp)
}

func (h *FounderHandler) Apply(c *gin.Context) {
	userID := mustCurrentUserID(c)
	var req dto.FounderApplyProfileReq
	if !BindValidate(c, &req) {
		return
	}
	fp, err := h.svc.Apply(c.Request.Context(), userID, &req)
	if err != nil {
		failContent(c, err)
		return
	}
	response.OK(c, fp)
}
