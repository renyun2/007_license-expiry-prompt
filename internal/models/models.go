package models

import "time"

// Certificate 企业资质证书台账
type Certificate struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	Name              string    `gorm:"size:255;not null" json:"name"`
	Category          string    `gorm:"size:64;not null;index" json:"category"` // 营业执照/行业许可证/...
	IssuingOrg        string    `gorm:"size:255" json:"issuing_org"`
	CertNo            string    `gorm:"size:128;index" json:"cert_no"`
	IssueDate         time.Time `json:"issue_date"`
	ExpiryDate        time.Time `gorm:"index" json:"expiry_date"`
	Department        string    `gorm:"size:128" json:"department"`
	ResponsiblePerson string    `gorm:"size:128" json:"responsible_person"`
	IsCancelled       bool      `json:"is_cancelled"` // true 表示已注销
	ScanURL           string    `gorm:"size:512" json:"scan_url"`
	FrontImageURL     string    `gorm:"size:512" json:"front_image_url"`
	BackImageURL      string    `gorm:"size:512" json:"back_image_url"`
	Notes             string    `gorm:"type:text" json:"notes"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`

	// 仅查询时填充
	ComputedStatus string `gorm:"-" json:"computed_status"` // 有效/即将到期/已过期/已注销
	UrgencyDays    int    `gorm:"-" json:"urgency_days"`    // 距离到期天数，已过期为负
}

// ReminderConfig 每类证书提醒阈值与默认负责人
type ReminderConfig struct {
	ID                  uint   `gorm:"primaryKey" json:"id"`
	Category            string `gorm:"size:64;uniqueIndex;not null" json:"category"`
	DaysTier1           int    `json:"days_tier1"` // 一级提醒：默认 180
	DaysTier2           int    `json:"days_tier2"` // 90
	DaysTier3           int    `json:"days_tier3"` // 30
	DefaultResponsible  string `gorm:"size:128" json:"default_responsible"`
}

// RenewalProgress 续期/年检进展
type RenewalProgress string

const (
	ProgressPreparing    RenewalProgress = "材料准备中"
	ProgressSubmitted    RenewalProgress = "已提交"
	ProgressReviewing    RenewalProgress = "审核中"
	ProgressApproved     RenewalProgress = "已通过"
)

// RenewalApplication 续期申请
type RenewalApplication struct {
	ID                 uint            `gorm:"primaryKey" json:"id"`
	CertificateID      uint            `gorm:"index;not null" json:"certificate_id"`
	Applicant          string          `gorm:"size:128" json:"applicant"`
	ApplyTime          time.Time       `json:"apply_time"`
	MaterialsChecklist string          `gorm:"type:text" json:"materials_checklist"` // JSON 数组字符串
	ExpectedSubmitDate *time.Time      `json:"expected_submit_date"`
	Progress           RenewalProgress `gorm:"size:32" json:"progress"`
	NewCertNotes       string          `gorm:"type:text" json:"new_cert_notes"` // 通过后录入说明
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	Certificate        *Certificate    `gorm:"foreignKey:CertificateID;constraint:OnDelete:CASCADE" json:"certificate,omitempty"`
}

// AnnualInspectionRecord 历届年检记录
type AnnualInspectionRecord struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	CertificateID uint      `gorm:"index;not null" json:"certificate_id"`
	Year          int       `gorm:"index" json:"year"`
	Certificate   *Certificate `gorm:"foreignKey:CertificateID;constraint:OnDelete:CASCADE" json:"-"`
	RecordURL     string    `gorm:"size:512" json:"record_url"`
	Notes         string    `gorm:"type:text" json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
}

// FeeRecord 年检/续期费用
type FeeRecord struct {
	ID                   uint       `gorm:"primaryKey" json:"id"`
	CertificateID        uint       `gorm:"index;not null" json:"certificate_id"`
	RenewalApplicationID *uint      `gorm:"index" json:"renewal_application_id"`
	AdminFeeCents        int64      `json:"admin_fee_cents"`  // 分
	AgencyFeeCents       int64      `json:"agency_fee_cents"` // 分
	FeeDate              time.Time  `gorm:"index" json:"fee_date"`
	Description          string     `gorm:"size:255" json:"description"`
	CreatedAt            time.Time  `json:"created_at"`
	Certificate          *Certificate `gorm:"foreignKey:CertificateID;constraint:OnDelete:CASCADE" json:"certificate,omitempty"`
}

// TodoItem 到期待办（由定时逻辑或查询生成，持久化便于跟踪）
type TodoItem struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	CertificateID uint      `gorm:"index;not null" json:"certificate_id"`
	Certificate   *Certificate `gorm:"foreignKey:CertificateID;constraint:OnDelete:CASCADE" json:"-"`
	Assignee      string    `gorm:"size:128" json:"assignee"`
	Title         string    `gorm:"size:255" json:"title"`
	DueDate       time.Time `gorm:"index" json:"due_date"`
	Done          bool      `json:"done"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
