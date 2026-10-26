package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"meandproduct/internal/model/entity"
	"meandproduct/internal/repository"
)

var (
	ErrInviteCodeNotFound     = errors.New("invite code not found")
	ErrInviteCodeNotYours     = errors.New("invite code does not belong to you")
	ErrInviteCodeAlreadyUsed  = errors.New("invite code already used, cannot revoke")
	ErrInviteInvalidMaxUses   = errors.New("invalid max_uses, must be >= 1 and <= 10000")
	ErrInviteInvalidExpiresAt = errors.New("invalid expires_at, must be after now")
	ErrInviteInvalidCount     = errors.New("invalid count, must be >= 1 and <= 100")
)

// GenerateRandomInviteCode 随机 16 字节 → hex32；保证足够熵且符合 dto `[A-Za-z0-9_-]{1,32}` regex（hex 仅字母数字）
func GenerateRandomInviteCode() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Errorf("invite code rand read failed: %w", err))
	}
	return hex.EncodeToString(buf)
}

// InviteService 邀请码业务层：founder 自己生成/列表/作废；admin 全局列表/批量生成/批量作废
type InviteService struct {
	repo *repository.UserRepo
}

func NewInviteService(repo *repository.UserRepo) *InviteService {
	return &InviteService{repo: repo}
}

// FounderGenerate founder 单条生成邀请码；maxUses=0 默认为 1；expiresAt=零值默认 30 天后
func (s *InviteService) FounderGenerate(ctx context.Context, creatorID uint64, maxUses int, expiresAt time.Time) (*entity.InviteCode, error) {
	if maxUses <= 0 {
		maxUses = 1
	}
	if maxUses < 1 || maxUses > 10000 {
		return nil, ErrInviteInvalidMaxUses
	}
	if expiresAt.IsZero() {
		tmp := time.Now().Add(30 * 24 * time.Hour)
		expiresAt = tmp
	}
	if !expiresAt.After(time.Now()) {
		return nil, ErrInviteInvalidExpiresAt
	}

	ic := &entity.InviteCode{
		Code:      GenerateRandomInviteCode(),
		CreatedBy: creatorID,
		MaxUses:   maxUses,
		UsedCount: 0,
		Status:    entity.InviteActive,
		ExpiresAt: &expiresAt,
	}
	if err := s.repo.CreateInviteCode(ctx, nil, ic); err != nil {
		return nil, err
	}
	return ic, nil
}

// FounderList founder 自己的邀请码分页
func (s *InviteService) FounderList(ctx context.Context, creatorID uint64, page, pageSize int) ([]*entity.InviteCode, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.ListInvitesByCreator(ctx, creatorID, pageSize, offset)
}

// FounderRevoke founder 作废自己的一条邀请码；仅允许 used_count=0（未消耗的才能作废避免歧义）
func (s *InviteService) FounderRevoke(ctx context.Context, creatorID, inviteID uint64) error {
	ic, err := s.repo.GetInviteByID(ctx, inviteID)
	if err != nil {
		return err
	}
	if ic == nil {
		return ErrInviteCodeNotFound
	}
	if ic.CreatedBy != creatorID {
		return ErrInviteCodeNotYours
	}
	if ic.Status == entity.InviteUsed {
		return ErrInviteCodeAlreadyUsed
	}
	if ic.Status == entity.InviteRevoked {
		return nil
	}
	rows, err := s.repo.RevokeInviteByCreator(ctx, inviteID, creatorID)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrInviteCodeNotYours
	}
	return nil
}

// AdminList admin 全局邀请码分页；status=空 不按状态过滤
func (s *InviteService) AdminList(ctx context.Context, status entity.InviteCodeStatus, page, pageSize int) ([]*entity.InviteCode, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 500 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize
	return s.repo.ListGlobalInvites(ctx, status, pageSize, offset)
}

// AdminGenerateMany admin 批量生成邀请码；默认单邀请码 max=1 expires=30d；count 1~100
func (s *InviteService) AdminGenerateMany(ctx context.Context, adminID uint64, count, maxUses int, expiresAt time.Time) ([]*entity.InviteCode, error) {
	if count < 1 || count > 100 {
		return nil, ErrInviteInvalidCount
	}
	if maxUses <= 0 {
		maxUses = 1
	}
	if maxUses < 1 || maxUses > 10000 {
		return nil, ErrInviteInvalidMaxUses
	}
	if expiresAt.IsZero() {
		tmp := time.Now().Add(30 * 24 * time.Hour)
		expiresAt = tmp
	}
	if !expiresAt.After(time.Now()) {
		return nil, ErrInviteInvalidExpiresAt
	}

	out := make([]*entity.InviteCode, 0, count)
	for i := 0; i < count; i++ {
		ic := &entity.InviteCode{
			Code:      GenerateRandomInviteCode(),
			CreatedBy: adminID,
			MaxUses:   maxUses,
			UsedCount: 0,
			Status:    entity.InviteActive,
			ExpiresAt: &expiresAt,
		}
		if err := s.repo.CreateInviteCode(ctx, nil, ic); err != nil {
			return nil, err
		}
		out = append(out, ic)
	}
	return out, nil
}

// AdminRevokeMany admin 批量作废邀请码（去重、忽略已 revoked）；返回 rows 实际更新条数
func (s *InviteService) AdminRevokeMany(ctx context.Context, ids []uint64) (int64, error) {
	seen := make(map[uint64]struct{}, len(ids))
	uniq := make([]uint64, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniq = append(uniq, id)
	}
	return s.repo.RevokeInvitesByAdmin(ctx, uniq)
}
