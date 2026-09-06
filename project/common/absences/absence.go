package absencespg

import (
	"time"
)

type Absence struct {
	Id             string    `json:"id,omitempty" gorm:"primaryKey"`
	OrganizationId string    `json:"organization_id,omitempty" gorm:"not null;index"`
	ChildId        string    `json:"child_id,omitempty" gorm:"not null;index"`
	ReportedBy     string    `json:"reported_by,omitempty" gorm:"not null"`
	DateFrom       time.Time `json:"date_from,omitempty" gorm:"not null"`
	DateTo         time.Time `json:"date_to,omitempty" gorm:"not null"`
	Reason         string    `json:"reason,omitempty"`
	CreateTime     time.Time `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime     time.Time `json:"update_time,omitempty" gorm:"not null"`
}
