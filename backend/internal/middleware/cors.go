package middleware

import (
	"github.com/gin-gonic/gin"

	"meandproduct/internal/config"
)

// CORS 允许跨域。开发环境默认 *，生产通过 config.CORS.AllowOrigins 严格控制。
func CORS(cfg *config.Config) gin.HandlerFunc {
	origins := cfg.CORS.OriginsSlice()
	methods := cfg.CORS.MethodsSlice()
	headers := cfg.CORS.HeadersSlice()
	allowAny := len(origins) == 1 && origins[0] == "*"

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			// 非浏览器请求，跳过
			c.Next()
			return
		}

		// 选择返回的 Allow-Origin
		var allowOrigin string
		if allowAny {
			allowOrigin = "*"
		} else {
			for _, o := range origins {
				if o == origin {
					allowOrigin = origin
					break
				}
			}
		}
		if allowOrigin == "" {
			c.AbortWithStatus(403)
			return
		}

		c.Header("Access-Control-Allow-Origin", allowOrigin)
		c.Header("Access-Control-Allow-Methods", joinStr(methods, ","))
		c.Header("Access-Control-Allow-Headers", joinStr(headers, ","))
		if cfg.CORS.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Header("Access-Control-Max-Age", itoa(cfg.CORS.MaxAge))

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func joinStr(items []string, sep string) string {
	out := ""
	for i, s := range items {
		if i > 0 {
			out += sep
		}
		out += s
	}
	return out
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
