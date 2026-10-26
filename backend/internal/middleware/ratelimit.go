package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"meandproduct/internal/config"
	"meandproduct/internal/pkg/response"
)

// RateLimitGeneral 全局限流：300 req / IP / 1min（固定窗口）
func RateLimitGeneral(cfg *config.Config, rd *redis.Client) gin.HandlerFunc {
	return rateLimit(rd,
		"rate:general:",
		cfg.RateLimit.GeneralLimit,
		cfg.RateLimit.GeneralWindow,
		false,
	)
}

// RateLimitLogin 登录接口限流：60 req / IP / 5min —— 登录 handler 层单独引用
func RateLimitLogin(cfg *config.Config, rd *redis.Client) gin.HandlerFunc {
	return rateLimit(rd,
		"rate:login:",
		cfg.RateLimit.LoginLimit,
		cfg.RateLimit.LoginWindow,
		false,
	)
}

// RateLimitReport 举报接口限流：5 req / IP / 1h
func RateLimitReport(cfg *config.Config, rd *redis.Client) gin.HandlerFunc {
	return rateLimit(rd,
		"rate:report:",
		cfg.RateLimit.ReportLimit,
		cfg.RateLimit.ReportWindow,
		false,
	)
}

func rateLimit(rd *redis.Client, keyPrefix string, limit int, window time.Duration, _ bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet && c.Request.URL.Path == "/api/v1/health" {
			c.Next()
			return
		}
		ip := c.ClientIP()
		key := fmt.Sprintf("%s%s", keyPrefix, ip)

		ctx := context.Background()
		pipe := rd.TxPipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, window)
		if _, err := pipe.Exec(ctx); err != nil {
			// Redis 异常时放行，不阻塞业务
			c.Next()
			return
		}
		if incr.Val() > int64(limit) {
			response.Fail(c, http.StatusTooManyRequests, "too many requests")
			c.Abort()
			return
		}
		c.Next()
	}
}
