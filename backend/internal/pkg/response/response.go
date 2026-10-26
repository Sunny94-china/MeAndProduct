package response

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Resp struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// OK 返回成功响应
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Resp{
		Code:      0,
		Message:   "ok",
		Data:      data,
		RequestID: ensureRequestID(c),
	})
}

// Fail 返回业务失败响应
func Fail(c *gin.Context, httpCode int, message string) {
	c.AbortWithStatusJSON(httpCode, Resp{
		Code:      httpCode,
		Message:   message,
		RequestID: ensureRequestID(c),
	})
}

// FailWithDetail 带附加 detail 字段（调试用，生产慎用）
func FailWithDetail(c *gin.Context, httpCode int, message string, detail any) {
	c.AbortWithStatusJSON(httpCode, gin.H{
		"code":       httpCode,
		"message":    message,
		"detail":     detail,
		"request_id": ensureRequestID(c),
	})
}

func ensureRequestID(c *gin.Context) string {
	v, ok := c.Get("request_id")
	if ok {
		return v.(string)
	}
	buf := make([]byte, 8)
	_, _ = rand.Read(buf)
	id := "req_" + hex.EncodeToString(buf)
	c.Set("request_id", id)
	return id
}
