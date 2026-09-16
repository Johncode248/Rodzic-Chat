package eventspg

import "time"

type EventRegistration struct {
	Id      string `json:"id,omitempty" gorm:"primaryKey"`
	EventId string `json:"event_id,omitempty" gorm:"not null;uniqueIndex:idx_event_registration_unique"`
	// The uniqueIndex above (shared with ChildId below) is defense-in-depth
	// on top of the pg_advisory_xact_lock in events.Declare: it guarantees
	// at the DB level that one child can never hold two registration rows
	// for the same event, which is exactly what a double-tapped "będzie"
	// button or a retried request used to be able to create.
	ChildId      string             `json:"child_id,omitempty" gorm:"not null;uniqueIndex:idx_event_registration_unique"`
	ConfirmedBy  string             `json:"confirmed_by,omitempty" gorm:"not null"`
	Status       RegistrationStatus `json:"status,omitempty" gorm:"not null"`
	ConsentGiven bool               `json:"consent_given,omitempty"`
	CreateTime   time.Time          `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime   time.Time          `json:"update_time,omitempty" gorm:"not null"`
}
