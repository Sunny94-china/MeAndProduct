package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"meandproduct/internal/model/entity"
)

// UserRepo 用户 + 创业者名片 + 邀请码 相关数据库访问
type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// Tx 在事务中执行 fn
func (r *UserRepo) Tx(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

// ================================================================
// User
// ================================================================

// CreateUser 插入一条 user，成功后填充 u.ID / CreatedAt
func (r *UserRepo) CreateUser(ctx context.Context, tx *gorm.DB, u *entity.User) error {
	d := r.db
	if tx != nil {
		d = tx
	}
	return d.WithContext(ctx).Create(u).Error
}

// GetByUsername 根据用户名取用户（软删返回 nil, nil）
func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var u entity.User
	err := r.db.WithContext(ctx).Where("username = ?", username).Take(&u).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByEmail 按邮箱取
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var u entity.User
	err := r.db.WithContext(ctx).Where("email = ?", email).Take(&u).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByID 按 ID 取
func (r *UserRepo) GetByID(ctx context.Context, id uint64) (*entity.User, error) {
	var u entity.User
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&u).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// UsernameOrEmailExists 用户名或邮箱任一存在
func (r *UserRepo) UsernameOrEmailExists(ctx context.Context, username, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("username = ? OR email = ?", username, email).
		Count(&count).Error
	return count > 0, err
}

// UpdateLastLoginAt 更新最后登录时间
func (r *UserRepo) UpdateLastLoginAt(ctx context.Context, userID uint64, t time.Time) error {
	return r.db.WithContext(ctx).
		Model(&entity.User{ID: userID}).
		UpdateColumn("last_login_at", t).Error
}

// UpdateStatus 更新账户状态
func (r *UserRepo) UpdateStatus(ctx context.Context, userID uint64, s entity.UserStatus) error {
	return r.db.WithContext(ctx).
		Model(&entity.User{ID: userID}).
		UpdateColumn("status", s).Error
}

// UpdateInviteStatusByCreator 只能 creator 本人操作：把自己的某条邀请码 status 改为 revoked（安全 WHERE created_by=?）
// 返回 rowsAffected：=0 表示该邀请码不存在或不属于此 creator
func (r *UserRepo) RevokeInviteByCreator(ctx context.Context, id, createdBy uint64) (int64, error) {
	res := r.db.WithContext(ctx).
		Model(&entity.InviteCode{}).
		Where("id = ? AND created_by = ? AND status <> ?", id, createdBy, entity.InviteRevoked).
		UpdateColumn("status", entity.InviteRevoked)
	return res.RowsAffected, res.Error
}

// RevokeInvitesByAdmin admin 批量 revoke（支持多个 ID）
func (r *UserRepo) RevokeInvitesByAdmin(ctx context.Context, ids []uint64) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	res := r.db.WithContext(ctx).
		Model(&entity.InviteCode{}).
		Where("id IN ? AND status <> ?", ids, entity.InviteRevoked).
		UpdateColumn("status", entity.InviteRevoked)
	return res.RowsAffected, res.Error
}

// ListGlobalInvites admin 全局邀请码列表分页；statusFilter 空=不按 status 过滤
func (r *UserRepo) ListGlobalInvites(ctx context.Context, statusFilter entity.InviteCodeStatus, limit, offset int) ([]*entity.InviteCode, int64, error) {
	var total int64
	base := r.db.WithContext(ctx).Model(&entity.InviteCode{})
	if statusFilter != "" {
		base = base.Where("status = ?", statusFilter)
	}
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*entity.InviteCode
	if limit <= 0 {
		limit = 20
	}
	err := base.Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&list).Error
	return list, total, err
}

// ================================================================
// Founder Profile
// ================================================================

// CreateFounderProfile 插入创业者名片空壳（注册后自动建）
func (r *UserRepo) CreateFounderProfile(ctx context.Context, tx *gorm.DB, p *entity.FounderProfile) error {
	d := r.db
	if tx != nil {
		d = tx
	}
	return d.WithContext(ctx).Create(p).Error
}

// GetFounderProfileByUserID 按 user_id 找名片；未找到返回 nil, nil
func (r *UserRepo) GetFounderProfileByUserID(ctx context.Context, userID uint64) (*entity.FounderProfile, error) {
	var p entity.FounderProfile
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Take(&p).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// FounderSlugExists 检查 slug 是否已被占用
func (r *UserRepo) FounderSlugExists(ctx context.Context, slug string) (bool, error) {
	var c int64
	err := r.db.WithContext(ctx).Model(&entity.FounderProfile{}).Where("slug = ?", slug).Count(&c).Error
	return c > 0, err
}

// ================================================================
// Invite Code
// ================================================================

// CreateInviteCode 新建邀请码
func (r *UserRepo) CreateInviteCode(ctx context.Context, tx *gorm.DB, c *entity.InviteCode) error {
	d := r.db
	if tx != nil {
		d = tx
	}
	return d.WithContext(ctx).Create(c).Error
}

// GetInviteByCode 按 code 取邀请码；不存在返回 nil, nil
func (r *UserRepo) GetInviteByCode(ctx context.Context, code string) (*entity.InviteCode, error) {
	var ic entity.InviteCode
	err := r.db.WithContext(ctx).Where("code = ?", code).Take(&ic).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ic, nil
}

// GetInviteByID 按 ID 取
func (r *UserRepo) GetInviteByID(ctx context.Context, id uint64) (*entity.InviteCode, error) {
	var ic entity.InviteCode
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&ic).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ic, nil
}

// ListInvitesByCreator 取某创建者的邀请码列表（按 created_at desc）
func (r *UserRepo) ListInvitesByCreator(ctx context.Context, createdBy uint64, limit, offset int) ([]*entity.InviteCode, int64, error) {
	var total int64
	base := r.db.WithContext(ctx).Model(&entity.InviteCode{}).Where("created_by = ?", createdBy)
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*entity.InviteCode
	if limit <= 0 {
		limit = 20
	}
	err := base.Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&list).Error
	return list, total, err
}

// ConsumeInvite 原子使用一次邀请码：used_count+1；若达到 max_uses，再把 status=used + used_at=now。
// WHERE id=? AND status=active AND used_count<max_uses 保证并发安全。
// rowsAffected=0 表示不可用（已用完或不活跃）。
func (r *UserRepo) ConsumeInvite(ctx context.Context, tx *gorm.DB, id uint64, now time.Time) (int64, error) {
	d := r.db
	if tx != nil {
		d = tx
	}
	res := d.WithContext(ctx).
		Model(&entity.InviteCode{}).
		Where("id = ? AND status = ? AND used_count < max_uses", id, entity.InviteActive).
		UpdateColumn("used_count", gorm.Expr("used_count + 1"))
	if res.Error != nil {
		return 0, res.Error
	}
	if res.RowsAffected == 0 {
		return 0, nil
	}
	res2 := d.WithContext(ctx).
		Model(&entity.InviteCode{}).
		Where("id = ? AND used_count >= max_uses AND status <> ?", id, entity.InviteUsed).
		Updates(map[string]any{
			"status":  entity.InviteUsed,
			"used_at": now,
		})
	if res2.Error != nil {
		return res.RowsAffected, res2.Error
	}
	return res.RowsAffected, nil
}

// RevokeInvite 作废邀请码（admin 全局无限制用；建议鉴权后才使用）
func (r *UserRepo) RevokeInvite(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).
		Model(&entity.InviteCode{ID: id}).
		UpdateColumn("status", entity.InviteRevoked).Error
}

// GetByRole 按角色批量取用户（admin 审核 founder 列表、admin 计数 etc 都用）
func (r *UserRepo) GetByRole(ctx context.Context, role entity.UserRole, limit, offset int) ([]*entity.User, int64, error) {
	var total int64
	base := r.db.WithContext(ctx).Model(&entity.User{}).Where("role = ?", role)
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*entity.User
	if limit <= 0 {
		limit = 50
	}
	err := base.Order("id DESC").Limit(limit).Offset(offset).Find(&list).Error
	return list, total, err
}

// MarkInviteExpired 把已过期且 status=active 的邀请码批量标记为 expired
func (r *UserRepo) MarkInviteExpired(ctx context.Context, now time.Time) (int64, error) {
	res := r.db.WithContext(ctx).
		Model(&entity.InviteCode{}).
		Where("status = ? AND expires_at IS NOT NULL AND expires_at < ?", entity.InviteActive, now).
		UpdateColumn("status", entity.InviteExpired)
	return res.RowsAffected, res.Error
}
