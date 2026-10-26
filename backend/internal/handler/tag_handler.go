package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"meandproduct/internal/model/dto"
	"meandproduct/internal/model/vo"
	"meandproduct/internal/pkg/response"
	"meandproduct/internal/service"
)

type TagHandler struct {
	svc *service.TagService
}

func NewTagHandler(s *service.TagService) *TagHandler {
	return &TagHandler{svc: s}
}

func (h *TagHandler) PublicListActive(c *gin.Context) {
	list, err := h.svc.ListActive(c.Request.Context())
	if err != nil {
		failContent(c, err)
		return
	}
	response.OK(c, list)
}

func (h *TagHandler) PublicGetGrouped(c *gin.Context) {
	groups, err := h.svc.GetByTypeGrouped(c.Request.Context())
	if err != nil {
		failContent(c, err)
		return
	}
	response.OK(c, vo.TagGroupedResponse{Groups: groups})
}

func (h *TagHandler) AdminListAll(c *gin.Context) {
	list, err := h.svc.ListActive(c.Request.Context())
	if err != nil {
		failContent(c, err)
		return
	}
	response.OK(c, list)
}

func (h *TagHandler) AdminCreate(c *gin.Context) {
	var req dto.TagCreateReq
	if !BindValidate(c, &req) {
		return
	}
	t, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		failContent(c, err)
		return
	}
	response.OK(c, vo.TagVO{
		ID:           t.ID,
		Name:         t.Name,
		Slug:         t.Slug,
		Type:         t.Type,
		FounderCount: t.FounderCount,
		ProductCount: t.ProductCount,
		SortOrder:    t.SortOrder,
		IsActive:     t.IsActive,
	})
}

func (h *TagHandler) AdminUpdate(c *gin.Context) {
	id, err := parseUintID(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var req dto.TagUpdateReq
	if !BindValidate(c, &req) {
		return
	}
	t, err := h.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		failContent(c, err)
		return
	}
	response.OK(c, vo.TagVO{
		ID:           t.ID,
		Name:         t.Name,
		Slug:         t.Slug,
		Type:         t.Type,
		FounderCount: t.FounderCount,
		ProductCount: t.ProductCount,
		SortOrder:    t.SortOrder,
		IsActive:     t.IsActive,
	})
}

func (h *TagHandler) AdminDeactivate(c *gin.Context) {
	id, err := parseUintID(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.Deactivate(c.Request.Context(), id); err != nil {
		failContent(c, err)
		return
	}
	response.OK(c, gin.H{"deactivated": true})
}
