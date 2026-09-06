package eventspg

import (
	"time"
)

type Event struct {
	Id                   string     `json:"id,omitempty" gorm:"primaryKey"`
	OrganizationId       string     `json:"organization_id,omitempty" gorm:"not null;index"`
	GroupId              string     `json:"group_id,omitempty" gorm:"index"`
	Type                 EventType  `json:"type,omitempty" gorm:"not null"`
	Title                string     `json:"title,omitempty" gorm:"not null"`
	Description          string     `json:"description,omitempty"`
	RequiresConsent      bool       `json:"requires_consent,omitempty"`
	RequiresItem         string     `json:"requires_item,omitempty"`
	EventDate            time.Time  `json:"event_date,omitempty" gorm:"not null"`
	RegistrationDeadline *time.Time `json:"registration_deadline,omitempty"`
	CreatedBy            string     `json:"created_by,omitempty" gorm:"not null"`
	CreateTime           time.Time  `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime           time.Time  `json:"update_time,omitempty" gorm:"not null"`
}
