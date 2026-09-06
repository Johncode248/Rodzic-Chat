package registrationlinkspg

import "time"

// RegistrationLink is how a director lets parents in one group self-register
// their own account, instead of the director creating parent accounts
// one-by-one (unlike teachers, who are still added manually by the
// director — see project/common/authorization/invite.go). A link is scoped
// to a single Group, caps out at ChildrenCount*2 completed registrations
// (at most two parent accounts per child — mother + father), and expires
// 30 days after creation. "Revoking" a link is a hard delete of this row —
// there's no separate status field, every other state (expired, exhausted)
// is computed from ExpiresAt/UsedCount at read time, not stored.
type RegistrationLink struct {
	Id               string    `json:"id,omitempty" gorm:"primaryKey"`
	OrganizationId   string    `json:"organization_id,omitempty" gorm:"not null;index"`
	GroupId          string    `json:"group_id,omitempty" gorm:"not null;index"`
	Token            string    `json:"token,omitempty" gorm:"not null;uniqueIndex"`
	ChildrenCount    int       `json:"children_count,omitempty" gorm:"not null"`
	MaxRegistrations int       `json:"max_registrations,omitempty" gorm:"not null"`
	UsedCount        int       `json:"used_count,omitempty" gorm:"not null"`
	CreatedBy        string    `json:"created_by,omitempty" gorm:"not null"`
	ExpiresAt        time.Time `json:"expires_at,omitempty" gorm:"not null"`
	CreateTime       time.Time `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime       time.Time `json:"update_time,omitempty" gorm:"not null"`
}

// IsUsable reports whether the link can still accept a new registration —
// not expired and not at its registration cap.
func (l RegistrationLink) IsUsable(now time.Time) bool {
	return now.Before(l.ExpiresAt) && l.UsedCount < l.MaxRegistrations
}
