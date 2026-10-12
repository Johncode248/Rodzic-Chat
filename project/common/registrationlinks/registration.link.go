package registrationlinkspg

import (
	"time"

	groupspg "github.com/Johncode248/Rodzic-Chat/project/common/groups"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
)

// RegistrationLink is how a director lets an already-authenticated parent
// in one group self-attach their own account, instead of the director
// creating parent accounts one-by-one (unlike teachers, who are still
// added manually by the director — see
// project/common/authorization/invite.go). A link is scoped to a single
// Group, caps out at ChildrenCount*2 completed registrations (at most two
// parent accounts per child — mother + father), and expires 72 hours after
// creation. "Revoking" a link is a hard delete of this row — there's no
// separate status field, every other state (expired, exhausted) is
// computed from ExpiresAt/UsedCount at read time, not stored.
//
// PinHash is the sha256 digest (see library/pinhash) of an 8-digit PIN
// (see library/random.GeneratePin) -- the plaintext PIN is returned once,
// on creation, and never stored anywhere.
type RegistrationLink struct {
	Id               string                        `json:"id,omitempty" gorm:"primaryKey"`
	OrganizationId   string                        `json:"organization_id,omitempty" gorm:"not null;index"`
	Organization     *organizationspg.Organization `json:"-" gorm:"foreignKey:OrganizationId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	GroupId          string                        `json:"group_id,omitempty" gorm:"not null;index"`
	Group            *groupspg.Group               `json:"-" gorm:"foreignKey:GroupId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	PinHash          string                        `json:"-" gorm:"not null;uniqueIndex"`
	ChildrenCount    int                           `json:"children_count,omitempty" gorm:"not null"`
	MaxRegistrations int                           `json:"max_registrations,omitempty" gorm:"not null"`
	UsedCount        int                           `json:"used_count,omitempty" gorm:"not null"`
	// PublicDescription is shown to whoever enters the PIN (e.g. "for
	// Zosia's mom") -- safe to return to any authenticated parent looking
	// this invite up.
	PublicDescription string `json:"public_description,omitempty"`
	// InternalNote is visible only to org staff (the create/list
	// endpoints) -- never returned from the lookup endpoint a PIN-enterer
	// calls.
	InternalNote string    `json:"internal_note,omitempty"`
	CreatedBy    string    `json:"created_by,omitempty" gorm:"not null"`
	ExpiresAt    time.Time `json:"expires_at,omitempty" gorm:"not null"`
	CreateTime   time.Time `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime   time.Time `json:"update_time,omitempty" gorm:"not null"`
}

// IsUsable reports whether the link can still accept a new registration —
// not expired and not at its registration cap.
func (l RegistrationLink) IsUsable(now time.Time) bool {
	return now.Before(l.ExpiresAt) && l.UsedCount < l.MaxRegistrations
}
