package calendarpg

import "time"

// Item unifies two very different sources into one feed for the parent-facing calendar: a personal
// ChildEvent (parent-created; children and/or parents attend) and a placowka Event (staff-created,
// org/group-scoped) -- see project/service/internal/calendar.Get. Source distinguishes which one
// this came from ("personal" or "placowka"); only personal items are editable by the parent.
//
// A repeating personal event is expanded server-side: each occurrence is its own Item sharing the
// same Id, with its own StartTime/EndTime; Occurrence is its 0-based index in the series.
// EventStartTime/EventEndTime always carry the series' ORIGINAL start/end (what an edit form
// should show), and edit/delete act on the whole series.
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

	// Personal events only.
	ChildIds           []string   `json:"child_ids,omitempty"`
	ChildNames         []string   `json:"child_names,omitempty"`
	ParentIds          []string   `json:"parent_ids,omitempty"`
	ParentNames        []string   `json:"parent_names,omitempty"`
	AllDay             bool       `json:"all_day,omitempty"`
	Recurrence         string     `json:"recurrence,omitempty"`
	RecurrenceInterval int        `json:"recurrence_interval,omitempty"`
	RecurrenceUntil    *time.Time `json:"recurrence_until,omitempty"`
	Occurrence         int        `json:"occurrence,omitempty"`
	EventStartTime     time.Time  `json:"event_start_time,omitempty"`
	EventEndTime       *time.Time `json:"event_end_time,omitempty"`
}

// Person is an account taking part in the calendar: the caller ("me") or a co-parent.
type Person struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type GetResponse struct {
	Items []Item `json:"items,omitempty"`
	// Me and CoParents (other parents sharing at least one child) let clients offer "who attends"
	// choices and colour the legend without extra calls.
	Me        Person   `json:"me,omitempty"`
	CoParents []Person `json:"co_parents,omitempty"`
}

const (
	SourcePersonal = "personal"
	SourcePlacowka = "placowka"
)
