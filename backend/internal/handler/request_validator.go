package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"meandproduct/internal/model/dto"
	"meandproduct/internal/pkg/response"
)

// Validatable 具有 Sanitize + Validate 方法的 DTO（handler 通用）
type Validatable interface {
	Sanitize()
	Validate() map[string]string
}

// BindValidate 绑定 + sanitize + validate；失败直接 Fail 写响应并返回 false
func BindValidate(c *gin.Context, v any) bool {
	if err := c.ShouldBindBodyWith(v, binding.JSON); err != nil {
		// 同时兼容 query / form，但本项目 DTO 默认全 JSON
		if errQ := c.ShouldBind(v); errQ != nil {
			msg := "invalid request parameters"
			lower := strings.ToLower(err.Error())
			switch {
			case strings.Contains(lower, "unmarshal"):
				msg = "invalid JSON body"
			case strings.Contains(lower, "required"):
				msg = "missing required field"
			}
			response.FailWithDetail(c, http.StatusBadRequest, msg, err.Error())
			return false
		}
	}
	if val, ok := v.(Validatable); ok {
		val.Sanitize()
		if errs := val.Validate(); len(errs) > 0 {
			response.FailWithDetail(c, http.StatusUnprocessableEntity, "validation failed", errs)
			return false
		}
	}
	return true
}

// placeholder 引用 dto，防 go vet unused（实际 BindValidate 会在 handler 中使用）
var _ = dto.RegisterSelfReq{}
