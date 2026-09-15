package calendarpg

import "time"

// Item unifies two very different sources into one feed for the parent-facing
// calendar: a personal ChildEvent (parent-created, any child) and a
// placowka Event (staff-created, org/group-scoped) -- see
// project/service/internal/calendar.Get. Source distinguishes which one
// this came from ("personal" or "placowka"); only personal items are
// editable by the parent (placowka events are read-only here, same as
// everywhere else in the app).
type Item struct {
	Id          string     `json:"id,omitempty"`
	Source      string     `json:"source,omitempty"`
	ChildId     string     `json:"child_id,omitempty"`
	ChildName   string     `json:"child_name,omitempty"`
	Title       string     `json:"title,omitempty"`
	Description string     `json:"description,omitempty"`
	Location    string     `json:"location,omitempty"`
	StartTime   time.Time  `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
}

type GetResponse struct {
	Items []Item `json:"items,omitempty"`
}

const (
	SourcePersonal = "personal"
	SourcePlacowka = "placowka"
)
