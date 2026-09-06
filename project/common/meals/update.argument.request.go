package mealspg

import "time"

type UpdateRequest struct {
	DateFrom *time.Time `json:"date_from,omitempty"`
	DateTo   *time.Time `json:"date_to,omitempty"`
	Reason   *string    `json:"reason,omitempty"`
}
