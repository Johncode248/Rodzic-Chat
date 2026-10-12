package childeventspg

// CreateRequest/UpdateRequest carry timestamps as RFC3339 strings (Go's standard JSON time.Time
// layout), same convention as project/common/absences and project/common/events.
//
// Attendees: ChildIds (the legacy single ChildId is still accepted), IncludeMe (the caller
// attends themselves; default true) and ParentIds (co-parents -- validated server-side to share a
// child with the caller).
type CreateRequest struct {
	ChildId            string   `json:"child_id,omitempty"`
	ChildIds           []string `json:"child_ids,omitempty"`
	IncludeMe          *bool    `json:"include_me,omitempty"`
	ParentIds          []string `json:"parent_ids,omitempty"`
	Title              string   `json:"title,omitempty"`
	Description        string   `json:"description,omitempty"`
	Location           string   `json:"location,omitempty"`
	StartTime          string   `json:"start_time,omitempty"`
	EndTime            string   `json:"end_time,omitempty"`
	AllDay             bool     `json:"all_day,omitempty"`
	Recurrence         string   `json:"recurrence,omitempty"`
	RecurrenceInterval int      `json:"recurrence_interval,omitempty"`
	RecurrenceUntil    string   `json:"recurrence_until,omitempty"`
	TimeZone           string   `json:"time_zone,omitempty"`
}

// UpdateRequest is a partial edit: only non-nil fields change. An empty EndTime / RecurrenceUntil
// string clears that field; Recurrence "" turns repeating off.
type UpdateRequest struct {
	ChildIds           *[]string `json:"child_ids,omitempty"`
	IncludeMe          *bool     `json:"include_me,omitempty"`
	ParentIds          *[]string `json:"parent_ids,omitempty"`
	Title              *string   `json:"title,omitempty"`
	Description        *string   `json:"description,omitempty"`
	Location           *string   `json:"location,omitempty"`
	StartTime          *string   `json:"start_time,omitempty"`
	EndTime            *string   `json:"end_time,omitempty"`
	AllDay             *bool     `json:"all_day,omitempty"`
	Recurrence         *string   `json:"recurrence,omitempty"`
	RecurrenceInterval *int      `json:"recurrence_interval,omitempty"`
	RecurrenceUntil    *string   `json:"recurrence_until,omitempty"`
	TimeZone           *string   `json:"time_zone,omitempty"`
}

type ListResponse struct {
	Events []ChildEvent `json:"events,omitempty"`
}
