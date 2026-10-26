package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"meandproduct/internal/middleware"
	"meandproduct/internal/model/dto"
	"meandproduct/internal/model/vo"
	"meandproduct/internal/pkg/response"
	"meandproduct/internal/service"
)

type ProductHandler struct {
	svc *service.ProductService
}

func NewProductHandler(s *service.ProductService) *ProductHandler {
	return &ProductHandler{svc: s}
}

// ====== PUBLIC ======

func (h *ProductHandler) PublicListGlobal(c *gin.Context) {
	var req dto.ProductListFilterReq
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
	list, total, err := h.svc.ListGlobal(c.Request.Context(), req)
	if err != nil {
		failContent(c, err)
		return
	}
	response.OK(c, vo.ProductListResponse{
		Page: vo.PageVO{Total: total, Page: *req.Page, Size: *req.PageSize},
		List: list,
	})
}

func (h *ProductHandler) PublicListByFounder(c *gin.Context) {
	founderSlug := c.Param("slug")
	var viewerUserID *uint64
	if u := middleware.GetCurrentUser(c); u != nil {
		viewerUserID = &u.UserID
	}
	list, total, err := h.svc.ListByFounder(c.Request.Context(), founderSlug, viewerUserID)
	if err != nil {
		failContent(c, err)
		return
	}
	response.OK(c, vo.ProductListResponse{
		Page: vo.PageVO{Total: total, Page: 1, Size: len(list)},
		List: list,
	})
}

func (h *ProductHandler) PublicGetByFounderSlugs(c *gin.Context) {
	founderSlug := c.Param("slug")
	productSlug := c.Param("p_slug")
	var viewer *uint64
	if u := middleware.GetCurrentUser(c); u != nil {
		viewer = &u.UserID
	}
	p, err := h.svc.GetByFounderSlugs(c.Request.Context(), founderSlug, productSlug, false, viewer)
	if err != nil {
		failContent(c, err)
		return
	}
	response.OK(c, p)
}

// ====== FOUNDER / ADMIN ======

func (h *ProductHandler) ListMine(c *gin.Context) {
	userID := mustCurrentUserID(c)
	page, pageSize := 1, 100
	if n, err := strconv.Atoi(c.Query("page")); err == nil && n > 0 {
		page = n
	}
	if n, err := strconv.Atoi(c.Query("page_size")); err == nil && n > 0 && n <= 100 {
		pageSize = n
	}
	list, total, err := h.svc.ListMine(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		failContent(c, err)
		return
	}
	response.OK(c, vo.ProductListResponse{
		Page: vo.PageVO{Total: total, Page: page, Size: pageSize},
		List: list,
	})
}

func (h *ProductHandler) GetMine(c *gin.Context) {
	userID := mustCurrentUserID(c)
	role := mustCurrentUserRole(c)
	id, err := parseUintID(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	p, err := h.svc.GetMine(c.Request.Context(), userID, role, id)
	if err != nil {
		failContent(c, err)
		return
	}
	response.OK(c, p)
}

func (h *ProductHandler) Create(c *gin.Context) {
	userID := mustCurrentUserID(c)
	var req dto.ProductCreateReq
	if !BindValidate(c, &req) {
		return
	}
	p, err := h.svc.Create(c.Request.Context(), userID, &req)
	if err != nil {
		failContent(c, err)
		return
	}
	response.OK(c, p)
}

func (h *ProductHandler) Update(c *gin.Context) {
	userID := mustCurrentUserID(c)
	role := mustCurrentUserRole(c)
	id, err := parseUintID(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var req dto.ProductUpdateReq
	if !BindValidate(c, &req) {
		return
	}
	p, err := h.svc.Update(c.Request.Context(), userID, role, id, &req)
	if err != nil {
		failContent(c, err)
		return
	}
	response.OK(c, p)
}

func (h *ProductHandler) Delete(c *gin.Context) {
	userID := mustCurrentUserID(c)
	role := mustCurrentUserRole(c)
	id, err := parseUintID(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), userID, role, id); err != nil {
		failContent(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *ProductHandler) Reorder(c *gin.Context) {
	userID := mustCurrentUserID(c)
	var req dto.ProductReorderReq
	if !BindValidate(c, &req) {
		return
	}
	if err := h.svc.Reorder(c.Request.Context(), userID, &req); err != nil {
		failContent(c, err)
		return
	}
	response.OK(c, gin.H{"reordered": true})
}

func parseUintID(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}
