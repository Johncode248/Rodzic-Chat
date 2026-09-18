package childeventspg

import (
	"time"

	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
)

// ChildEvent is a personal calendar entry a parent creates for one of their
// children (a doctor's visit, a birthday party, anything) -- entirely
// independent of any placowka. Unlike project/common/events.Event
// (organization-scoped, staff-managed), this works for every child,
// personal or placowka-connected alike, and any parent linked to the child
// (see project/service/internal/parentchild.IsParent) can create/edit/
// delete it -- there's no single "owner" beyond CreatedByAccountId, which
// is informational only.
type ChildEvent struct {
	Id                 string            `json:"id,omitempty" gorm:"primaryKey"`
	ChildId            string            `json:"child_id,omitempty" gorm:"not null;index"`
	Child              *childrenpg.Child `json:"-" gorm:"foreignKey:ChildId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	CreatedByAccountId string            `json:"created_by_account_id,omitempty" gorm:"not null;index"`
	Title              string            `json:"title,omitempty" gorm:"not null"`
	Description        string            `json:"description,omitempty"`
	Location           string            `json:"location,omitempty"`
	StartTime          time.Time         `json:"start_time,omitempty" gorm:"not null;index"`
	EndTime            *time.Time        `json:"end_time,omitempty"`
	CreateTime         time.Time         `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime         time.Time         `json:"update_time,omitempty" gorm:"not null"`
}
