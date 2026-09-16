package accountspg

import (
	"time"

	"gorm.io/gorm"
)

type Account struct {
	Id             string      `json:"id,omitempty" gorm:"primaryKey"`
	OrganizationId string      `json:"organization_id,omitempty" gorm:"index"`
	Name           string      `json:"name,omitempty" gorm:"not null"`
	Login          string      `json:"login,omitempty" gorm:"unique; not null"`
	Email          string      `json:"email,omitempty" gorm:"index"`
	Phone          string      `json:"phone,omitempty"`
	AccountType    AccountType `json:"account_type,omitempty" gorm:"not null"`
	Status         Status      `json:"status,omitempty" gorm:"not null"`
	// Permissions grants a Teacher account specific extra capabilities
	// (e.g. PermissionManagePayments) beyond what AccountType alone allows.
	// Always 0 for Parent accounts; irrelevant for Director/CoDirector,
	// which implicitly have every permission everywhere it's checked.
	Permissions Permission `json:"permissions,omitempty" gorm:"not null"`
	CreateTime  time.Time  `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime  time.Time  `json:"update_time,omitempty" gorm:"not null"`
	// DeletedAt is GORM's built-in soft-delete convention: every existing
	// gormdb.Select/gormdb.Delete call site (there are dozens) automatically
	// excludes/soft-deletes on this field with zero code changes, because
	// they all go through the generic gormdb helpers which just call
	// tx.Find/tx.Delete and let GORM detect the field. The 3-month reaper
	// (cmd/reaper) is the only code that needs to bypass this via Unscoped().
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}
