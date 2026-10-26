package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"meandproduct/internal/config"
	"meandproduct/internal/middleware"
	"meandproduct/internal/model/dto"
	"meandproduct/internal/model/vo"
	"meandproduct/internal/pkg/jwtutil"
	"meandproduct/internal/pkg/response"
	"meandproduct/internal/service"
)

// AuthHandler 认证模块 HTTP Handler
type AuthHandler struct {
	cfg *config.Config
	svc *service.AuthService
	rd  *redis.Client
}

func NewAuthHandler(cfg *config.Config, svc *service.AuthService, rd *redis.Client) *AuthHandler {
	return &AuthHandler{cfg: cfg, svc: svc, rd: rd}
}

// ---- TokenPairResp（和 service TokenPair + MeVO 合成） ----
type TokenPairResp struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
	User             *vo.MeVO  `json:"user,omitempty"`
}

// errMap 把 service 业务错误映射到 HTTP 码
func errHTTP(err error) (int, string) {
	switch {
	case err == nil:
		return 0, ""
	case errors.Is(err, service.ErrConflictUsernameEmail):
		return http.StatusConflict, "username or email already exists"
	case errors.Is(err, service.ErrInvalidCredentials):
		return http.StatusUnauthorized, "invalid username/email or password"
	case errors.Is(err, service.ErrUserSuspendedOrPending):
		return http.StatusForbidden, "account suspended or pending, please contact admin"
	case errors.Is(err, service.ErrInviteCodeInvalid):
		return http.StatusUnprocessableEntity, "invalid or unavailable invite code"
	case errors.Is(err, service.ErrRefreshTokenInvalid):
		return http.StatusUnauthorized, "invalid or revoked refresh token"
	case errors.Is(err, service.ErrUserNotFound):
		return http.StatusNotFound, "user not found"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

// ---- RegisterSelf ----
func (h *AuthHandler) RegisterSelf(c *gin.Context) {
	var req dto.RegisterSelfReq
	if !BindValidate(c, &req) {
		return
	}
	pair, _, _, err := h.svc.RegisterSelf(c.Request.Context(), req.Username, req.Email, req.Password, req.DisplayName)
	if err != nil {
		code, msg := errHTTP(err)
		if code == http.StatusInternalServerError {
			c.Error(err)
		}
		response.Fail(c, code, msg)
		return
	}
	me, _ := h.svc.Me(c.Request.Context(), parseUserIDFromTokens(h.cfg, pair.AccessToken))
	response.OK(c, TokenPairResp{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		AccessExpiresAt:  pair.AccessExpiresAt,
		RefreshExpiresAt: pair.RefreshExpiresAt,
		User:             me,
	})
}

// ---- RegisterInvite ----
func (h *AuthHandler) RegisterInvite(c *gin.Context) {
	var req dto.RegisterInviteReq
	if !BindValidate(c, &req) {
		return
	}
	pair, _, _, err := h.svc.RegisterInvite(c.Request.Context(), req.Username, req.Email, req.Password, req.DisplayName, req.InviteCode)
	if err != nil {
		code, msg := errHTTP(err)
		if code == http.StatusInternalServerError {
			c.Error(err)
		}
		response.Fail(c, code, msg)
		return
	}
	me, _ := h.svc.Me(c.Request.Context(), parseUserIDFromTokens(h.cfg, pair.AccessToken))
	response.OK(c, TokenPairResp{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		AccessExpiresAt:  pair.AccessExpiresAt,
		RefreshExpiresAt: pair.RefreshExpiresAt,
		User:             me,
	})
}

// ---- Login ----
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginReq
	if !BindValidate(c, &req) {
		return
	}
	pair, u, err := h.svc.Login(c.Request.Context(), req.UsernameOrEmail, req.Password)
	if err != nil {
		code, msg := errHTTP(err)
		if code == http.StatusInternalServerError {
			c.Error(err)
		}
		response.Fail(c, code, msg)
		return
	}
	me, _ := h.svc.Me(c.Request.Context(), u.ID)
	response.OK(c, TokenPairResp{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		AccessExpiresAt:  pair.AccessExpiresAt,
		RefreshExpiresAt: pair.RefreshExpiresAt,
		User:             me,
	})
}

// ---- Refresh ----
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshReq
	if !BindValidate(c, &req) {
		return
	}
	pair, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		code, msg := errHTTP(err)
		if code == http.StatusInternalServerError {
			c.Error(err)
		}
		response.Fail(c, code, msg)
		return
	}
	me, _ := h.svc.Me(c.Request.Context(), parseUserIDFromTokens(h.cfg, pair.AccessToken))
	response.OK(c, TokenPairResp{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		AccessExpiresAt:  pair.AccessExpiresAt,
		RefreshExpiresAt: pair.RefreshExpiresAt,
		User:             me,
	})
}

// ---- Logout ----
func (h *AuthHandler) Logout(c *gin.Context) {
	cu := middleware.GetCurrentUser(c)
	if cu == nil {
		response.Fail(c, http.StatusUnauthorized, "login required")
		return
	}
	var req dto.LogoutReq
	// body 可选（缺省时只拉黑 access）
	_ = BindValidate(c, &req)
	now := time.Now()
	// 用 access 的 claims.Exp 算剩余 TTL 也行：直接用 access TTL 替代，足够安全
	accessExp := now.Add(h.cfg.JWT.AccessTTL)
	refreshExp := now.Add(h.cfg.JWT.RefreshTTL)
	refreshJti := ""
	if req.RefreshToken != "" {
		rc, err := jwtutil.Parse(h.cfg, req.RefreshToken, jwtutil.TokenTypeRefresh)
		if err == nil {
			refreshJti = rc.JWTID
			refreshExp = time.Unix(rc.Exp, 0)
		}
	}
	if err := h.svc.Logout(c.Request.Context(), cu.Jti, refreshJti, accessExp, refreshExp); err != nil {
		response.Fail(c, http.StatusInternalServerError, "logout failed")
		return
	}
	response.OK(c, gin.H{"logout": true})
}

// ---- ValidateInvite ----
func (h *AuthHandler) ValidateInvite(c *gin.Context) {
	var req dto.ValidateInviteReq
	if !BindValidate(c, &req) {
		return
	}
	res, err := h.svc.ValidateInvite(c.Request.Context(), req.InviteCode)
	if err != nil {
		c.Error(err)
		response.Fail(c, http.StatusInternalServerError, "internal error")
		return
	}
	response.OK(c, res)
}

// ---- RegisterGuest ----
func (h *AuthHandler) RegisterGuest(c *gin.Context) {
	var req dto.RegisterSelfReq
	if !BindValidate(c, &req) {
		return
	}
	pair, user, err := h.svc.RegisterGuest(c.Request.Context(), req.Username, req.Email, req.Password, req.DisplayName)
	if err != nil {
		code, msg := errHTTP(err)
		if code == http.StatusInternalServerError {
			c.Error(err)
		}
		response.Fail(c, code, msg)
		return
	}
	me, _ := h.svc.Me(c.Request.Context(), user.ID)
	response.OK(c, TokenPairResp{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		AccessExpiresAt:  pair.AccessExpiresAt,
		RefreshExpiresAt: pair.RefreshExpiresAt,
		User:             me,
	})
}

// ---- Me ----
func (h *AuthHandler) Me(c *gin.Context) {
	cu := middleware.GetCurrentUser(c)
	if cu == nil {
		response.Fail(c, http.StatusUnauthorized, "login required")
		return
	}
	me, err := h.svc.Me(c.Request.Context(), cu.UserID)
	if err != nil {
		code, msg := errHTTP(err)
		if code == http.StatusInternalServerError {
			c.Error(err)
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, me)
}

// parseUserIDFromTokens 内部用：注册/登录/refresh 后查 /auth/me 相同信息
func parseUserIDFromTokens(cfg *config.Config, accessToken string) uint64 {
	c, err := jwtutil.Parse(cfg, accessToken, jwtutil.TokenTypeAccess)
	if err != nil || c == nil {
		return 0
	}
	return c.UserID
}
