package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"meandproduct/internal/config"
	"meandproduct/internal/model/entity"
	"meandproduct/internal/model/vo"
	"meandproduct/internal/pkg/jwtutil"
	"meandproduct/internal/pkg/password"
	"meandproduct/internal/pkg/slug"
	"meandproduct/internal/repository"
)

// 业务错误码常量（handler 映射 HTTP 状态）
var (
	ErrConflictUsernameEmail  = errors.New("username or email already exists")
	ErrInvalidCredentials     = errors.New("invalid username/email or password")
	ErrUserSuspendedOrPending = errors.New("account is suspended or pending, cannot login")
	ErrInviteCodeInvalid      = errors.New("invite code is invalid or unavailable")
	ErrRefreshTokenInvalid    = errors.New("refresh token invalid or revoked")
	ErrUserNotFound           = errors.New("user not found")
)

// TokenPair 对外部响应的 token 对 VO（薄封装）
type TokenPair struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

// InviteUsable 邀请码可用性查询结果
type InviteUsable struct {
	Usable        bool       `json:"usable"`
	Message       string     `json:"message"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	RemainingUses int        `json:"remaining_uses"`
}

// AuthService 认证与用户注册/登录业务层
type AuthService struct {
	cfg  *config.Config
	db   *gorm.DB
	rd   *redis.Client
	repo *repository.UserRepo
}

func NewAuthService(cfg *config.Config, db *gorm.DB, rd *redis.Client, repo *repository.UserRepo) *AuthService {
	return &AuthService{cfg: cfg, db: db, rd: rd, repo: repo}
}

// EnsureConfiguredAdmin 按配置写入固定管理员。已存在则同步密码、角色和状态。
func EnsureConfiguredAdmin(ctx context.Context, cfg *config.Config, db *gorm.DB) error {
	username := strings.TrimSpace(strings.ToLower(cfg.Admin.Username))
	email := strings.TrimSpace(strings.ToLower(cfg.Admin.Email))
	rawPwd := cfg.Admin.Password
	if username == "" || email == "" || rawPwd == "" {
		return nil
	}
	hash, err := password.Hash(rawPwd)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}
	repo := repository.NewUserRepo(db)
	existing, err := repo.GetByUsername(ctx, username)
	if err != nil {
		return err
	}
	if existing == nil {
		byEmail, err := repo.GetByEmail(ctx, email)
		if err != nil {
			return err
		}
		if byEmail != nil {
			return fmt.Errorf("admin email %s already used by %s", email, byEmail.Username)
		}
		u := &entity.User{
			Username:       username,
			Email:          email,
			PasswordHash:   hash,
			Role:           entity.RoleAdmin,
			Status:         entity.UserStatusActive,
			RegisterSource: entity.SourceSelf,
		}
		if err := repo.CreateUser(ctx, nil, u); err != nil {
			return fmt.Errorf("create admin: %w", err)
		}
		return nil
	}
	return db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", existing.ID).Updates(map[string]any{
		"email":         email,
		"password_hash": hash,
		"role":          entity.RoleAdmin,
		"status":        entity.UserStatusActive,
	}).Error
}

// RegisterSelf 自助注册（创建 founder 身份 + 待审核 founder profile）
// 登录规则：允许 pending founder 登录（仅 founder-dashboard 自助 Apply 补资料、看自己的 Profile），前台浏览不依赖 founder 登录
func (a *AuthService) RegisterSelf(ctx context.Context, username, email, rawPwd, displayName string) (*TokenPair, *entity.User, *entity.FounderProfile, error) {
	username, email = strings.TrimSpace(strings.ToLower(username)), strings.TrimSpace(strings.ToLower(email))
	if displayName == "" {
		displayName = username
	}
	exists, err := a.repo.UsernameOrEmailExists(ctx, username, email)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("repo UsernameOrEmailExists: %w", err)
	}
	if exists {
		return nil, nil, nil, ErrConflictUsernameEmail
	}
	hash, err := password.Hash(rawPwd)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("password hash: %w", err)
	}
	var (
		createdUser    *entity.User
		createdFounder *entity.FounderProfile
		pair           *TokenPair
	)
	err = a.repo.Tx(ctx, func(tx *gorm.DB) error {
		u := &entity.User{
			Username:       username,
			Email:          email,
			PasswordHash:   hash,
			Role:           entity.RoleFounder,
			Status:         entity.UserStatusPending,
			RegisterSource: entity.SourceSelf,
		}
		if err := a.repo.CreateUser(ctx, tx, u); err != nil {
			return fmt.Errorf("CreateUser: %w", err)
		}
		createdUser = u

		baseSlug := slug.Generate("", username)
		uniqueSlug, err := slug.IsUnique(baseSlug, func(s string) (bool, error) {
			return a.repo.FounderSlugExists(ctx, s)
		})
		if err != nil {
			return fmt.Errorf("slug unique: %w", err)
		}
		fp := &entity.FounderProfile{
			UserID:      u.ID,
			DisplayName: displayName,
			Slug:        uniqueSlug,
			IsVisible:   false,
		}
		if err := a.repo.CreateFounderProfile(ctx, tx, fp); err != nil {
			return fmt.Errorf("CreateFounderProfile: %w", err)
		}
		createdFounder = fp
		tp, err := jwtutil.IssueTokens(a.cfg, u.ID, string(u.Role))
		if err != nil {
			return fmt.Errorf("IssueTokens: %w", err)
		}
		pair = &TokenPair{
			AccessToken:      tp.AccessToken.Token,
			RefreshToken:     tp.RefreshToken.Token,
			AccessExpiresAt:  tp.AccessToken.ExpiresAt,
			RefreshExpiresAt: tp.RefreshToken.ExpiresAt,
		}
		return nil
	})
	if err != nil {
		return nil, nil, nil, err
	}
	return pair, createdUser, createdFounder, nil
}

// RegisterGuest 浏览用户注册（role=guest，status=active，不创建 founder profile，可立即登录收藏等前台操作）
func (a *AuthService) RegisterGuest(ctx context.Context, username, email, rawPwd, displayName string) (*TokenPair, *entity.User, error) {
	username, email = strings.TrimSpace(strings.ToLower(username)), strings.TrimSpace(strings.ToLower(email))
	if displayName == "" {
		displayName = username
	}
	exists, err := a.repo.UsernameOrEmailExists(ctx, username, email)
	if err != nil {
		return nil, nil, fmt.Errorf("repo UsernameOrEmailExists: %w", err)
	}
	if exists {
		return nil, nil, ErrConflictUsernameEmail
	}
	hash, err := password.Hash(rawPwd)
	if err != nil {
		return nil, nil, fmt.Errorf("password hash: %w", err)
	}
	var created *entity.User
	var pair *TokenPair
	err = a.repo.Tx(ctx, func(tx *gorm.DB) error {
		u := &entity.User{
			Username:       username,
			Email:          email,
			PasswordHash:   hash,
			Role:           entity.RoleGuest,
			Status:         entity.UserStatusActive,
			RegisterSource: entity.SourceSelf,
		}
		if err := a.repo.CreateUser(ctx, tx, u); err != nil {
			return fmt.Errorf("CreateUser: %w", err)
		}
		created = u
		tp, err := jwtutil.IssueTokens(a.cfg, u.ID, string(u.Role))
		if err != nil {
			return fmt.Errorf("IssueTokens: %w", err)
		}
		pair = &TokenPair{
			AccessToken:      tp.AccessToken.Token,
			RefreshToken:     tp.RefreshToken.Token,
			AccessExpiresAt:  tp.AccessToken.ExpiresAt,
			RefreshExpiresAt: tp.RefreshToken.ExpiresAt,
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return pair, created, nil
}

// RegisterInvite 使用邀请码注册（status=active，is_visible=true，ConsumeInvite 原子扣减）
func (a *AuthService) RegisterInvite(ctx context.Context, username, email, rawPwd, displayName, code string) (*TokenPair, *entity.User, *entity.FounderProfile, error) {
	username, email, code = strings.TrimSpace(strings.ToLower(username)), strings.TrimSpace(strings.ToLower(email)), strings.TrimSpace(code)
	if displayName == "" {
		displayName = username
	}
	ic, err := a.repo.GetInviteByCode(ctx, code)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("GetInviteByCode: %w", err)
	}
	if ic == nil || !ic.Usable(time.Now()) {
		return nil, nil, nil, ErrInviteCodeInvalid
	}
	exists, err := a.repo.UsernameOrEmailExists(ctx, username, email)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("UsernameOrEmailExists: %w", err)
	}
	if exists {
		return nil, nil, nil, ErrConflictUsernameEmail
	}
	hash, err := password.Hash(rawPwd)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("password hash: %w", err)
	}
	var (
		u        *entity.User
		fp       *entity.FounderProfile
		pair     *TokenPair
		inviteID = ic.ID
	)
	err = a.repo.Tx(ctx, func(tx *gorm.DB) error {
		n, err := a.repo.ConsumeInvite(ctx, tx, inviteID, time.Now())
		if err != nil {
			return fmt.Errorf("ConsumeInvite: %w", err)
		}
		if n == 0 {
			return ErrInviteCodeInvalid
		}
		us := &entity.User{
			Username:       username,
			Email:          email,
			PasswordHash:   hash,
			Role:           entity.RoleFounder,
			Status:         entity.UserStatusActive,
			RegisterSource: entity.SourceInvite,
			InviteCodeID:   &inviteID,
		}
		if err := a.repo.CreateUser(ctx, tx, us); err != nil {
			return fmt.Errorf("CreateUser: %w", err)
		}
		u = us
		baseSlug := slug.Generate("", username)
		uniqueSlug, err := slug.IsUnique(baseSlug, func(s string) (bool, error) {
			return a.repo.FounderSlugExists(ctx, s)
		})
		if err != nil {
			return fmt.Errorf("slug unique: %w", err)
		}
		fpObj := &entity.FounderProfile{
			UserID:      u.ID,
			DisplayName: displayName,
			Slug:        uniqueSlug,
			IsVisible:   true,
		}
		if err := a.repo.CreateFounderProfile(ctx, tx, fpObj); err != nil {
			return fmt.Errorf("CreateFounderProfile: %w", err)
		}
		fp = fpObj
		tp, err := jwtutil.IssueTokens(a.cfg, u.ID, string(u.Role))
		if err != nil {
			return fmt.Errorf("IssueTokens: %w", err)
		}
		pair = &TokenPair{
			AccessToken:      tp.AccessToken.Token,
			RefreshToken:     tp.RefreshToken.Token,
			AccessExpiresAt:  tp.AccessToken.ExpiresAt,
			RefreshExpiresAt: tp.RefreshToken.ExpiresAt,
		}
		return nil
	})
	if err != nil {
		return nil, nil, nil, err
	}
	return pair, u, fp, nil
}

// Login 登录：username 或 email 登录，返回新 token 对
func (a *AuthService) Login(ctx context.Context, usernameOrEmail, rawPwd string) (*TokenPair, *entity.User, error) {
	usernameOrEmail = strings.TrimSpace(usernameOrEmail)
	var u *entity.User
	// 先按用户名再按邮箱
	got, err := a.repo.GetByUsername(ctx, strings.ToLower(usernameOrEmail))
	if err != nil {
		return nil, nil, fmt.Errorf("GetByUsername: %w", err)
	}
	if got != nil {
		u = got
	} else {
		byEmail, err := a.repo.GetByEmail(ctx, strings.ToLower(usernameOrEmail))
		if err != nil {
			return nil, nil, fmt.Errorf("GetByEmail: %w", err)
		}
		u = byEmail
	}
	if u == nil {
		_ = password.Verify(rawPwd, "$2a$14$CFCFCFCFCFCFCFCFCFCFCFu5OqR9vFfX5U4Y7p0yS5U4Y7p0yS5U4Y7p") // dummy
		return nil, nil, ErrInvalidCredentials
	}
	if err := password.Verify(rawPwd, u.PasswordHash); err != nil {
		return nil, nil, ErrInvalidCredentials
	}
	if !u.CanLogin() && u.Status != entity.UserStatusPending {
		return nil, nil, ErrUserSuspendedOrPending
	}
	_ = a.repo.UpdateLastLoginAt(ctx, u.ID, time.Now())
	tp, err := jwtutil.IssueTokens(a.cfg, u.ID, string(u.Role))
	if err != nil {
		return nil, nil, fmt.Errorf("IssueTokens: %w", err)
	}
	return &TokenPair{
		AccessToken:      tp.AccessToken.Token,
		RefreshToken:     tp.RefreshToken.Token,
		AccessExpiresAt:  tp.AccessToken.ExpiresAt,
		RefreshExpiresAt: tp.RefreshToken.ExpiresAt,
	}, u, nil
}

// Refresh 用 refresh 换出新 pair；旧 refresh jti 立即拉黑
func (a *AuthService) Refresh(ctx context.Context, refreshTokenStr string) (*TokenPair, error) {
	refreshTokenStr = strings.TrimSpace(refreshTokenStr)
	if refreshTokenStr == "" {
		return nil, ErrRefreshTokenInvalid
	}
	claims, err := jwtutil.Parse(a.cfg, refreshTokenStr, jwtutil.TokenTypeRefresh)
	if err != nil {
		return nil, ErrRefreshTokenInvalid
	}
	// 检查旧 refresh 是否已拉黑
	blacklisted, err := jwtutil.IsBlacklisted(ctx, a.cfg, a.rd, claims.JWTID)
	if err != nil {
		return nil, fmt.Errorf("IsBlacklisted: %w", err)
	}
	if blacklisted {
		return nil, ErrRefreshTokenInvalid
	}
	// 旧 refresh jti 拉黑（剩余 TTL）
	remain := time.Until(time.Unix(claims.Exp, 0))
	if remain > 0 {
		_ = jwtutil.Blacklist(ctx, a.cfg, a.rd, claims.JWTID, remain)
	}
	// 用户存在性检查
	u, err := a.repo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	if u == nil || !u.CanLogin() {
		return nil, ErrRefreshTokenInvalid
	}
	tp, err := jwtutil.IssueTokens(a.cfg, u.ID, string(u.Role))
	if err != nil {
		return nil, fmt.Errorf("IssueTokens: %w", err)
	}
	return &TokenPair{
		AccessToken:      tp.AccessToken.Token,
		RefreshToken:     tp.RefreshToken.Token,
		AccessExpiresAt:  tp.AccessToken.ExpiresAt,
		RefreshExpiresAt: tp.RefreshToken.ExpiresAt,
	}, nil
}

// Logout 拉黑当前 access + refresh（剩余 TTL 作为 Redis TTL）
func (a *AuthService) Logout(ctx context.Context, accessJTI, refreshJTI string, accessExp, refreshExp time.Time) error {
	now := time.Now()
	if accessJTI != "" {
		ttl := accessExp.Sub(now)
		if ttl > 0 {
			if err := jwtutil.Blacklist(ctx, a.cfg, a.rd, accessJTI, ttl); err != nil {
				return fmt.Errorf("blacklist access: %w", err)
			}
		}
	}
	if refreshJTI != "" {
		ttl := refreshExp.Sub(now)
		if ttl > 0 {
			if err := jwtutil.Blacklist(ctx, a.cfg, a.rd, refreshJTI, ttl); err != nil {
				return fmt.Errorf("blacklist refresh: %w", err)
			}
		}
	}
	return nil
}

// ValidateInvite 邀请码可用性查询（仅读）
func (a *AuthService) ValidateInvite(ctx context.Context, code string) (*InviteUsable, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return &InviteUsable{Usable: false, Message: "code is required"}, nil
	}
	ic, err := a.repo.GetInviteByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("GetInviteByCode: %w", err)
	}
	if ic == nil {
		return &InviteUsable{Usable: false, Message: "invite code not found"}, nil
	}
	now := time.Now()
	switch {
	case ic.Status == entity.InviteRevoked:
		return &InviteUsable{Usable: false, Message: "invite code revoked", ExpiresAt: ic.ExpiresAt}, nil
	case ic.Status == entity.InviteUsed:
		return &InviteUsable{Usable: false, Message: "invite code fully used", ExpiresAt: ic.ExpiresAt}, nil
	case ic.Status == entity.InviteExpired:
		return &InviteUsable{Usable: false, Message: "invite code expired", ExpiresAt: ic.ExpiresAt}, nil
	case ic.ExpiresAt != nil && now.After(*ic.ExpiresAt):
		return &InviteUsable{Usable: false, Message: "invite code expired", ExpiresAt: ic.ExpiresAt}, nil
	case ic.UsedCount >= ic.MaxUses:
		return &InviteUsable{Usable: false, Message: "invite code fully used", ExpiresAt: ic.ExpiresAt}, nil
	}
	return &InviteUsable{
		Usable:        true,
		Message:       "ok",
		ExpiresAt:     ic.ExpiresAt,
		RemainingUses: ic.MaxUses - ic.UsedCount,
	}, nil
}

// Me 查询当前用户信息
func (a *AuthService) Me(ctx context.Context, userID uint64) (*vo.MeVO, error) {
	u, err := a.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	me := &vo.MeVO{
		ID:             u.ID,
		Username:       u.Username,
		Email:          u.Email,
		Role:           string(u.Role),
		Status:         string(u.Status),
		RegisterSource: string(u.RegisterSource),
		AvatarURL:      u.AvatarURL,
		LastLoginAt:    u.LastLoginAt,
		CreatedAt:      u.CreatedAt,
	}
	if u.Role == entity.RoleFounder {
		fp, err := a.repo.GetFounderProfileByUserID(ctx, u.ID)
		if err != nil {
			return nil, fmt.Errorf("GetFounderProfileByUserID: %w", err)
		}
		if fp != nil {
			customs, _ := vo.ParseCustomLinks(fp.CustomLinks)
			fv := &vo.FounderProfileVO{
				ID:            fp.ID,
				UserID:        fp.UserID,
				DisplayName:   fp.DisplayName,
				Slug:          fp.Slug,
				IsVisible:     fp.IsVisible,
				CustomLinks:   customs,
				ViewCount:     fp.ViewCount,
				FavoriteCount: fp.FavoriteCount,
				CreatedAt:     fp.CreatedAt,
				UpdatedAt:     fp.UpdatedAt,
			}
			// is_visible=true 或 active 状态时返回完整字段
			if fp.IsVisible || u.Status == entity.UserStatusActive {
				fv.Title = fp.Title
				fv.Bio = fp.Bio
				fv.BioLong = fp.BioLong
				fv.Location = fp.Location
				fv.AvatarURL = fp.AvatarURL
				fv.CoverURL = fp.CoverURL
				fv.ContactEmail = fp.ContactEmail
				fv.WebsiteURL = fp.WebsiteURL
				fv.GithubURL = fp.GithubURL
				fv.TwitterURL = fp.TwitterURL
				fv.LinkedinURL = fp.LinkedinURL
				fv.ZhihuURL = fp.ZhihuURL
				fv.XiaohongshuURL = fp.XiaohongshuURL
				fv.ProducthuntURL = fp.ProducthuntURL
				fv.WechatID = fp.WechatID
			}
			me.Founder = fv
		}
	}
	return me, nil
}
