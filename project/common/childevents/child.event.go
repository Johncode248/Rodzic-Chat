package childeventspg

import (
	"time"

	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
)

// Recurrence frequencies a personal event can repeat with ("" = does not repeat).
const (
	RecurrenceNone    = ""
	RecurrenceDaily   = "daily"
	RecurrenceWeekly  = "weekly"
	RecurrenceMonthly = "monthly"
	RecurrenceYearly  = "yearly"
)

// ChildEvent is a personal calendar entry a parent creates -- for one or more of their children,
// for themselves, and/or for a co-parent (a doctor's visit, a birthday party, a parent-teacher
// meeting, anything) -- entirely independent of any placowka. Unlike project/common/events.Event
// (organization-scoped, staff-managed), it works for personal and placowka-connected children
// alike, and any parent taking part in it (an attending parent or a parent of an attending child)
// may edit or delete it.
//
// It can span several days (StartTime..EndTime) and repeat (Recurrence + RecurrenceInterval,
// optionally until RecurrenceUntil). A repeating event is ONE row; the calendar feed expands its
// occurrences (see calendar.Get), and edit/delete act on the whole series.
//
// Attendees live in two join tables (ChildEventChild / ChildEventParent). ChildId is the legacy
// "primary child" column from when an event belonged to exactly one child; it is kept (first child,
// or "" for a parent-only event) and no longer has a foreign key.
type ChildEvent struct {
	Id                 string     `json:"id,omitempty" gorm:"primaryKey"`
	ChildId            string     `json:"child_id,omitempty" gorm:"index"`
	CreatedByAccountId string     `json:"created_by_account_id,omitempty" gorm:"not null;index"`
	Title              string     `json:"title,omitempty" gorm:"not null"`
	Description        string     `json:"description,omitempty"`
	Location           string     `json:"location,omitempty"`
	StartTime          time.Time  `json:"start_time,omitempty" gorm:"not null;index"`
	EndTime            *time.Time `json:"end_time,omitempty"`
	AllDay             bool       `json:"all_day,omitempty" gorm:"not null;default:false"`
	Recurrence         string     `json:"recurrence,omitempty"`
	// TimeZone is the IANA zone (e.g. "Europe/Warsaw") the event was scheduled in; a repeating event
	// is expanded in it so it keeps its wall-clock time across daylight-saving changes.
	TimeZone           string     `json:"time_zone,omitempty"`
	RecurrenceInterval int        `json:"recurrence_interval,omitempty" gorm:"not null;default:1"`
	RecurrenceUntil    *time.Time `json:"recurrence_until,omitempty"`
	CreateTime         time.Time  `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime         time.Time  `json:"update_time,omitempty" gorm:"not null"`

	// Computed on read, never persisted.
	ChildIds  []string `json:"child_ids,omitempty" gorm:"-"`
	ParentIds []string `json:"parent_ids,omitempty" gorm:"-"`
}

// ChildEventChild links an event to a child taking part in it.
type ChildEventChild struct {
	EventId string            `json:"event_id,omitempty" gorm:"primaryKey;index"`
	ChildId string            `json:"child_id,omitempty" gorm:"primaryKey;index"`
	Event   *ChildEvent       `json:"-" gorm:"foreignKey:EventId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Child   *childrenpg.Child `json:"-" gorm:"foreignKey:ChildId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
}

func (ChildEventChild) TableName() string { return "child_event_children" }

// ChildEventParent links an event to a parent account taking part in it (the creator, and/or a
// co-parent -- another parent linked to at least one of the same children).
type ChildEventParent struct {
	EventId   string              `json:"event_id,omitempty" gorm:"primaryKey;index"`
	AccountId string              `json:"account_id,omitempty" gorm:"primaryKey;index"`
	Event     *ChildEvent         `json:"-" gorm:"foreignKey:EventId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Account   *accountspg.Account `json:"-" gorm:"foreignKey:AccountId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
}

func (ChildEventParent) TableName() string { return "child_event_parents" }
