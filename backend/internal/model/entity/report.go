package entity

import "time"

type ReportReasonType string

const (
	ReasonSpam          ReportReasonType = "spam"
	ReasonFraud         ReportReasonType = "fraud"
	ReasonInappropriate ReportReasonType = "inappropriate"
	ReasonCopyright     ReportReasonType = "copyright"
	ReasonOther         ReportReasonType = "other"
)

type ReportStatus string

const (
	ReportPending  ReportStatus = "pending"
	ReportReviewed ReportStatus = "reviewed"
	ReportRejected ReportStatus = "rejected"
	ReportResolved ReportStatus = "resolved"
)

type ReportTarget string

const (
	ReportTargetFounder ReportTarget = "founder"
	ReportTargetProduct ReportTarget = "product"
)

type Report struct {
	ID           uint64           `gorm:"primaryKey;autoIncrement;column:id"`
	ReporterID   *uint64          `gorm:"column:reporter_id;index:idx_reporter"`
	TargetType   ReportTarget     `gorm:"column:target_type;size:20;not null;index:idx_target,priority:1"`
	TargetID     uint64           `gorm:"column:target_id;not null;index:idx_target,priority:2"`
	ReasonType   ReportReasonType `gorm:"column:reason_type;size:20;not null"`
	ReasonDetail string           `gorm:"column:reason_detail;type:text;not null"`
	ContactInfo  *string          `gorm:"column:contact_info;size:200"`
	Status       ReportStatus     `gorm:"column:status;size:20;not null;default:pending;index:idx_status"`
	HandlerID    *uint64          `gorm:"column:handler_id"`
	HandlerNote  *string          `gorm:"column:handler_note;type:text"`
	HideTarget   bool             `gorm:"column:hide_target;not null;default:false"`
	HandledAt    *time.Time       `gorm:"column:handled_at"`
	CreatedAt    time.Time        `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time        `gorm:"column:updated_at;autoUpdateTime"`
}

func (Report) TableName() string { return "reports" }
