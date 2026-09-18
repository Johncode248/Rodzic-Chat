package eventspg

import "time"

type EventUpdateRequest struct {
	Title                *string    `json:"title,omitempty"`
	Description          *string    `json:"description,omitempty"`
	RequiresItem         *string    `json:"requires_item,omitempty"`
	EventDate            *time.Time `json:"event_date,omitempty"`
	RegistrationDeadline *time.Time `json:"registration_deadline,omitempty"`
}
