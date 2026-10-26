package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"meandproduct/internal/model/dto"
	"meandproduct/internal/model/entity"
	"meandproduct/internal/model/vo"
	"meandproduct/internal/pkg/response"
	"meandproduct/internal/service"
)

type FavoriteHandler struct {
	svc *service.FavoriteService
}

func NewFavoriteHandler(s *service.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{svc: s}
}

func (h *FavoriteHandler) Toggle(c *gin.Context) {
	userID := mustCurrentUserID(c)
	var req dto.FavoriteToggleReq
	if !BindValidate(c, &req) {
		return
	}
	res, err := h.svc.Toggle(c.Request.Context(), userID, &req)
	if err != nil {
		failContent(c, err)
		return
	}
	response.OK(c, res)
}

func (h *FavoriteHandler) List(c *gin.Context) {
	userID := mustCurrentUserID(c)
	var req dto.FavoriteListReq
	if !BindValidate(c, &req) {
		return
	}
	if req.Page == nil {
		v := 1
		req.Page = &v
	}
	if req.PageSize == nil {
		v := 20
		req.PageSize = &v
	}
	if *req.Page < 1 {
		v := 1
		req.Page = &v
	}
	if *req.PageSize < 1 || *req.PageSize > 100 {
		v := 20
		req.PageSize = &v
	}
	list, total, err := h.svc.ListByUser(c.Request.Context(), userID, req)
	if err != nil {
		failContent(c, err)
		return
	}
	response.OK(c, vo.FavoriteListResponse{
		Page: vo.PageVO{Total: total, Page: *req.Page, Size: *req.PageSize},
		List: list,
	})
}

func (h *FavoriteHandler) Exists(c *gin.Context) {
	userID := mustCurrentUserID(c)
	targetType := entity.FavoriteTarget(c.Param("target_type"))
	idStr := c.Param("target_id")
	var id uint64
	if idStr != "" {
		n, err := strconv.ParseUint(idStr, 10, 64)
		if err == nil {
			id = n
		}
	}
	req := dto.FavoriteToggleReq{TargetType: targetType, TargetID: id}
	errs := req.Validate()
	if len(errs) > 0 {
		for _, v := range errs {
			response.Fail(c, 400, v)
			return
		}
	}
	ok, err := h.svc.Exists(c.Request.Context(), userID, &req)
	if err != nil {
		failContent(c, err)
		return
	}
	favCount := uint32(0)
	fp, _ := h.svc.SnapshotCount(c.Request.Context(), req.TargetType, req.TargetID)
	if fp != nil {
		switch req.TargetType {
		case entity.FavoriteTargetFounder:
			favCount = fp.FavoriteCount
		case entity.FavoriteTargetProduct:
			favCount = fp.FavoriteCount
		}
	}
	response.OK(c, gin.H{"favorited": ok, "favorite_count": favCount})
}
