package mealspg

import "time"

type CreateRequest struct {
	ChildId  string    `json:"child_id,omitempty"`
	DateFrom time.Time `json:"date_from,omitempty"`
	DateTo   time.Time `json:"date_to,omitempty"`
	Reason   string    `json:"reason,omitempty"`
}
