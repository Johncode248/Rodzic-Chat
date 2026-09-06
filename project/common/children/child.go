package childrenpg

import (
	"time"
)

type Child struct {
	Id             string    `json:"id,omitempty" gorm:"primaryKey"`
	OrganizationId string    `json:"organization_id,omitempty" gorm:"not null;index"`
	GroupId        string    `json:"group_id,omitempty" gorm:"index"`
	FirstName      string    `json:"first_name,omitempty" gorm:"not null"`
	LastName       string    `json:"last_name,omitempty" gorm:"not null"`
	BirthDate      time.Time `json:"birth_date,omitempty"`
	Notes          string    `json:"notes,omitempty"`
	CreateTime     time.Time `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime     time.Time `json:"update_time,omitempty" gorm:"not null"`
}
