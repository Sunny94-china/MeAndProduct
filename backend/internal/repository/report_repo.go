package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"meandproduct/internal/model/entity"
)

type ReportRepo struct {
	db *gorm.DB
}

func NewReportRepo(db *gorm.DB) *ReportRepo {
	return &ReportRepo{db: db}
}

type ReportListFilter struct {
	Status     entity.ReportStatus
	ReporterID *uint64
	TargetType entity.ReportTarget
	Page       int
	PageSize   int
}

func (r *ReportRepo) Create(ctx context.Context, tx *gorm.DB, rp *entity.Report) error {
	d := r.db
	if tx != nil {
		d = tx
	}
	return d.WithContext(ctx).Create(rp).Error
}

func (r *ReportRepo) GetByID(ctx context.Context, id uint64) (*entity.Report, error) {
	var rp entity.Report
	if err := r.db.WithContext(ctx).Where("id = ?", id).Take(&rp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &rp, nil
}

func (r *ReportRepo) List(ctx context.Context, filter ReportListFilter) ([]*entity.Report, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 500 {
		filter.PageSize = 20
	}
	q := r.db.WithContext(ctx).Model(&entity.Report{})
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.ReporterID != nil {
		q = q.Where("reporter_id = ?", *filter.ReporterID)
	}
	if filter.TargetType != "" {
		q = q.Where("target_type = ?", filter.TargetType)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (filter.Page - 1) * filter.PageSize
	items := make([]*entity.Report, 0, filter.PageSize)
	if err := q.Order("created_at DESC, id DESC").
		Offset(offset).Limit(filter.PageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

type HandleReportPayload struct {
	HandlerID  uint64
	NewStatus  entity.ReportStatus
	HandlerNote *string
	HideTarget  bool
}

func (r *ReportRepo) Handle(ctx context.Context, tx *gorm.DB, id uint64, p HandleReportPayload) (targetType entity.ReportTarget, targetID uint64, err error) {
	d := r.db
	if tx != nil {
		d = tx
	}
	updates := map[string]any{
		"status":      p.NewStatus,
		"handler_id":  p.HandlerID,
		"handler_note": p.HandlerNote,
		"handled_at":  time.Now(),
	}
	if err = d.WithContext(ctx).Model(&entity.Report{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return "", 0, err
	}
	var rp entity.Report
	if err = d.WithContext(ctx).Where("id = ?", id).Select("target_type, target_id").Take(&rp).Error; err != nil {
		return "", 0, err
	}
	targetType = rp.TargetType
	targetID = rp.TargetID
	if !p.HideTarget || targetID == 0 {
		return
	}
	switch rp.TargetType {
	case entity.ReportTargetFounder:
		err = d.WithContext(ctx).Model(&entity.FounderProfile{}).
			Where("id = ?", rp.TargetID).UpdateColumn("is_visible", false).Error
	case entity.ReportTargetProduct:
		err = d.WithContext(ctx).Model(&entity.Product{}).
			Where("id = ?", rp.TargetID).UpdateColumn("is_visible", false).Error
	}
	return
}
