package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"meandproduct/internal/config"
	"meandproduct/internal/pkg/jwtutil"
	"meandproduct/internal/pkg/response"
)

// CurrentUserKey 存入 gin context 的 key
const CurrentUserKey = "currentUser"

type CurrentUser struct {
	UserID uint64 `json:"user_id"`
	Role   string `json:"role"`
	Jti    string `json:"jti"`
}

// JWTAuth 解析 Bearer access token，校验通过注入 currentUser
func JWTAuth(cfg *config.Config, rd *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		if raw == "" || !strings.HasPrefix(raw, "Bearer ") {
			response.Fail(c, http.StatusUnauthorized, "missing authorization header")
			return
		}
		tokenStr := strings.TrimPrefix(raw, "Bearer ")

		claims, err := jwtutil.Parse(cfg, tokenStr, jwtutil.TokenTypeAccess)
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, "invalid or expired access token")
			return
		}

		black, err := jwtutil.IsBlacklisted(c.Request.Context(), cfg, rd, claims.JWTID)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "auth service unavailable")
			return
		}
		if black {
			response.Fail(c, http.StatusUnauthorized, "token revoked")
			return
		}

		u := &CurrentUser{UserID: claims.UserID, Role: claims.Role, Jti: claims.JWTID}
		c.Set(CurrentUserKey, u)
		c.Next()
	}
}

// RBAC 基于角色的权限控制。admin 隐含拥有所有其它角色。
// 传 "guest" 表示允许任何角色（也允许未登录？不 —— 这里 guest 表示：已登录且 role 符合要求。要允许未登录，不要在路由链路上挂载这个中间件。）
func RBAC(allowedRoles ...string) gin.HandlerFunc {
	roleIndex := make(map[string]struct{}, len(allowedRoles))
	for _, r := range allowedRoles {
		roleIndex[r] = struct{}{}
	}
	// admin 永久放行
	roleIndex["admin"] = struct{}{}

	return func(c *gin.Context) {
		v, ok := c.Get(CurrentUserKey)
		if !ok {
			response.Fail(c, http.StatusUnauthorized, "login required")
			return
		}
		u, ok := v.(*CurrentUser)
		if !ok {
			response.Fail(c, http.StatusForbidden, "invalid current user")
			return
		}
		if _, ok := roleIndex[u.Role]; !ok {
			response.Fail(c, http.StatusForbidden, "insufficient permission")
			return
		}
		c.Next()
	}
}

// GetCurrentUser 从 context 中取出当前用户，如果未设置返回 nil
func GetCurrentUser(c *gin.Context) *CurrentUser {
	v, ok := c.Get(CurrentUserKey)
	if !ok {
		return nil
	}
	u, _ := v.(*CurrentUser)
	return u
}

// Helper 保留未使用，防止 go vet 报错
var _ = time.Now
