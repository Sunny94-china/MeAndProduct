package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"meandproduct/internal/middleware"
	"meandproduct/internal/model/dto"
	"meandproduct/internal/model/vo"
	"meandproduct/internal/pkg/response"
	"meandproduct/internal/service"
)

type ReportHandler struct {
	svc *service.ReportService
}

func NewReportHandler(s *service.ReportService) *ReportHandler {
	return &ReportHandler{svc: s}
}

func (h *ReportHandler) Submit(c *gin.Context) {
	var reporterID *uint64
	if u := middleware.GetCurrentUser(c); u != nil {
		reporterID = &u.UserID
	}
	var req dto.ReportSubmitReq
	if !BindValidate(c, &req) {
		return
	}
	r, err := h.svc.Create(c.Request.Context(), reporterID, &req)
	if err != nil {
		failContent(c, err)
		return
	}
	response.OK(c, r)
}

func (h *ReportHandler) AdminList(c *gin.Context) {
	var req dto.ReportListFilterReq
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
	response.OK(c, vo.ReportListResponse{
		Page: vo.PageVO{Total: total, Page: *req.Page, Size: *req.PageSize},
		List: list,
	})
}

func (h *ReportHandler) AdminHandle(c *gin.Context) {
	id, err := parseUintID(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	handlerID := mustCurrentUserID(c)
	var req dto.ReportHandleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		msg := "invalid request parameters"
		lower := strings.ToLower(err.Error())
		switch {
		case strings.Contains(lower, "unmarshal"):
			msg = "invalid JSON body"
		case strings.Contains(lower, "required"):
			msg = "missing required field"
		}
		response.FailWithDetail(c, http.StatusBadRequest, msg, err.Error())
		return
	}
	req.ReportID = id
	if val, ok := any(&req).(Validatable); ok {
		val.Sanitize()
		if errs := val.Validate(); len(errs) > 0 {
			response.FailWithDetail(c, http.StatusUnprocessableEntity, "validation failed", errs)
			return
		}
	}
	r, err := h.svc.Handle(c.Request.Context(), handlerID, &req)
	if err != nil {
		failContent(c, err)
		return
	}
	response.OK(c, r)
}
