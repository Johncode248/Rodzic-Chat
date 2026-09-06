package eventspg

import "time"

type EventRegistration struct {
	Id           string             `json:"id,omitempty" gorm:"primaryKey"`
	EventId      string             `json:"event_id,omitempty" gorm:"not null;index"`
	ChildId      string             `json:"child_id,omitempty" gorm:"not null;index"`
	ConfirmedBy  string             `json:"confirmed_by,omitempty" gorm:"not null"`
	Status       RegistrationStatus `json:"status,omitempty" gorm:"not null"`
	ConsentGiven bool               `json:"consent_given,omitempty"`
	CreateTime   time.Time          `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime   time.Time          `json:"update_time,omitempty" gorm:"not null"`
}
