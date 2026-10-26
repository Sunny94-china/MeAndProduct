package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"meandproduct/internal/model/vo"
	"meandproduct/internal/pkg/response"
	"meandproduct/internal/service"
)

type LinkHandler struct {
	svc *service.LinkService
}

func NewLinkHandler(s *service.LinkService) *LinkHandler {
	return &LinkHandler{svc: s}
}

func (h *LinkHandler) TrackClick(c *gin.Context) {
	id, err := parseUintID(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	link, err := h.svc.TrackClick(c.Request.Context(), id)
	if err != nil {
		failContent(c, err)
		return
	}
	response.OK(c, vo.CustomLinkVO{
		ID:         link.ID,
		Label:      link.LinkType,
		LinkType:   link.LinkType,
		URL:        link.URL,
		ClickCount: link.ClickCount,
	})
}
