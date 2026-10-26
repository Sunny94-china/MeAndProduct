package entity

import (
	"database/sql"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// UserRole 用户角色：guest（浏览用户）/ founder（创业者）/ admin（管理员）
type UserRole string

const (
	RoleGuest   UserRole = "guest"
	RoleFounder UserRole = "founder"
	RoleAdmin   UserRole = "admin"
)

// UserStatus 账户状态
type UserStatus string

const (
	UserStatusPending   UserStatus = "pending"   // 自助申请待审核
	UserStatusActive    UserStatus = "active"    // 正常
	UserStatusSuspended UserStatus = "suspended" // 封禁/停用
)

// RegisterSource 注册来源
type RegisterSource string

const (
	SourceSelf   RegisterSource = "self"
	SourceInvite RegisterSource = "invite"
)

// User 用户账户表实体（users）
type User struct {
	ID             uint64         `gorm:"primaryKey;autoIncrement;column:id"`
	Username       string         `gorm:"size:50;not null;uniqueIndex:uk_username"`
	Email          string         `gorm:"size:100;not null;uniqueIndex:uk_email"`
	PasswordHash   string         `gorm:"column:password_hash;size:255;not null"`
	Role           UserRole       `gorm:"size:20;not null;default:founder;index:idx_role_status,priority:1"`
	Status         UserStatus     `gorm:"size:20;not null;default:pending;index:idx_role_status,priority:2"`
	AvatarURL      *string        `gorm:"column:avatar_url;size:500;default:null"`
	RegisterSource RegisterSource `gorm:"column:register_source;size:20;not null"`
	InviteCodeID   *uint64        `gorm:"column:invite_code_id;default:null"`
	LastLoginAt    *time.Time     `gorm:"column:last_login_at;default:null"`
	CreatedAt      time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index;default:null"`
}

func (User) TableName() string { return "users" }

// CanLogin 检查该账户是否允许登录。pending/suspended 不允许。
func (u *User) CanLogin() bool {
	if u == nil {
		return false
	}
	return u.Status == UserStatusActive
}

// FounderProfile 创业者名片（founder_profiles）
type FounderProfile struct {
	ID             uint64         `gorm:"primaryKey;autoIncrement;column:id"`
	UserID         uint64         `gorm:"column:user_id;not null;uniqueIndex:uk_user_id"`
	DisplayName    string         `gorm:"column:display_name;size:100;not null"`
	Slug           string         `gorm:"column:slug;size:100;not null;uniqueIndex:uk_slug"`
	Title          *string        `gorm:"column:title;size:100"`
	Bio            *string        `gorm:"column:bio;size:500"`
	BioLong        *string        `gorm:"column:bio_long;type:text"`
	Location       *string        `gorm:"column:location;size:100"`
	AvatarURL      *string        `gorm:"column:avatar_url;size:500"`
	CoverURL       *string        `gorm:"column:cover_url;size:500"`
	ContactEmail   *string        `gorm:"column:contact_email;size:100"`
	WebsiteURL     *string        `gorm:"column:website_url;size:500"`
	GithubURL      *string        `gorm:"column:github_url;size:500"`
	TwitterURL     *string        `gorm:"column:twitter_url;size:500"`
	LinkedinURL    *string        `gorm:"column:linkedin_url;size:500"`
	WechatID       *string        `gorm:"column:wechat_id;size:100"`
	ZhihuURL       *string        `gorm:"column:zhihu_url;size:500"`
	XiaohongshuURL *string        `gorm:"column:xiaohongshu_url;size:500"`
	ProducthuntURL *string        `gorm:"column:producthunt_url;size:500"`
	CustomLinks    sql.NullString `gorm:"column:custom_links;type:json"` // JSON array of {name,url}
	ViewCount      uint32         `gorm:"column:view_count;not null;default:0;index:idx_view_count,sort:desc"`
	FavoriteCount  uint32         `gorm:"column:favorite_count;not null;default:0;index:idx_favorite_count,sort:desc"`
	IsVisible      bool           `gorm:"column:is_visible;not null;default:true;index:idx_visible"`
	CreatedAt      time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index;default:null"`
}

func (FounderProfile) TableName() string { return "founder_profiles" }

// SetCustomLinks 把传入的 links 切片编码为 JSON 存入字段。
func (fp *FounderProfile) SetCustomLinks(links any) error {
	if links == nil {
		fp.CustomLinks = sql.NullString{Valid: false}
		return nil
	}
	b, err := json.Marshal(links)
	if err != nil {
		return err
	}
	fp.CustomLinks = sql.NullString{String: string(b), Valid: true}
	return nil
}

// InviteCodeStatus 邀请码状态
type InviteCodeStatus string

const (
	InviteActive  InviteCodeStatus = "active"
	InviteExpired InviteCodeStatus = "expired"
	InviteUsed    InviteCodeStatus = "used"
	InviteRevoked InviteCodeStatus = "revoked"
)

// InviteCode 邀请码（invite_codes）
type InviteCode struct {
	ID        uint64           `gorm:"primaryKey;autoIncrement;column:id"`
	Code      string           `gorm:"column:code;size:32;not null;uniqueIndex:uk_code"`
	CreatedBy uint64           `gorm:"column:created_by;not null;index:idx_created_by"`
	MaxUses   int              `gorm:"column:max_uses;not null;default:1"`
	UsedCount int              `gorm:"column:used_count;not null;default:0"`
	ExpiresAt *time.Time       `gorm:"column:expires_at;default:null"`
	Status    InviteCodeStatus `gorm:"column:status;size:20;not null;default:active;index:idx_status"`
	CreatedAt time.Time        `gorm:"column:created_at;autoCreateTime"`
	UsedAt    *time.Time       `gorm:"column:used_at;default:null"`
}

func (InviteCode) TableName() string { return "invite_codes" }

// Usable 当前是否可被使用。true=active & 未过期 & 还有名额
func (ic *InviteCode) Usable(now time.Time) bool {
	if ic == nil {
		return false
	}
	if ic.Status != InviteActive {
		return false
	}
	if ic.ExpiresAt != nil && now.After(*ic.ExpiresAt) {
		return false
	}
	return ic.UsedCount < ic.MaxUses
}
