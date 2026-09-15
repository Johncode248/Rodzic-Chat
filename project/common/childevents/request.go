package childeventspg

// CreateRequest/UpdateRequest carry timestamps as RFC3339 strings (Go's
// standard JSON time.Time layout), same convention as
// project/common/absences and project/common/events elsewhere in this
// codebase.
type CreateRequest struct {
	ChildId     string `json:"child_id,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Location    string `json:"location,omitempty"`
	StartTime   string `json:"start_time,omitempty"`
	EndTime     string `json:"end_time,omitempty"`
}

type UpdateRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Location    *string `json:"location,omitempty"`
	StartTime   *string `json:"start_time,omitempty"`
	EndTime     *string `json:"end_time,omitempty"`
}

type ListResponse struct {
	Events []ChildEvent `json:"events,omitempty"`
}
