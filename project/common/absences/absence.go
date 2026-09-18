package absencespg

import (
	"time"

	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
)

type Absence struct {
	Id             string                        `json:"id,omitempty" gorm:"primaryKey"`
	OrganizationId string                        `json:"organization_id,omitempty" gorm:"not null;index"`
	Organization   *organizationspg.Organization `json:"-" gorm:"foreignKey:OrganizationId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	ChildId        string                        `json:"child_id,omitempty" gorm:"not null;index"`
	Child          *childrenpg.Child             `json:"-" gorm:"foreignKey:ChildId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	ReportedBy     string                        `json:"reported_by,omitempty" gorm:"not null"`
	DateFrom       time.Time                     `json:"date_from,omitempty" gorm:"not null"`
	DateTo         time.Time                     `json:"date_to,omitempty" gorm:"not null"`
	Reason         string                        `json:"reason,omitempty"`
	CreateTime     time.Time                     `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime     time.Time                     `json:"update_time,omitempty" gorm:"not null"`
}
