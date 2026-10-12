package teacherinvitespg

import (
	"time"

	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
)

// TeacherInvite is how a director invites a teacher (or, per AccountType,
// a co-director): a single-use, expiring, revocable link — mirroring
// project/common/registrationlinks' parent flow instead of the old "create
// the account immediately" model, so a director can see pending invites
// and cancel one before it's used. GroupId is optional: a director can
// invite a teacher before deciding which group they'll lead (meaningless
// for a co-director invite, which is never group-scoped).
type TeacherInvite struct {
	Id             string                        `json:"id,omitempty" gorm:"primaryKey"`
	OrganizationId string                        `json:"organization_id,omitempty" gorm:"not null;index"`
	Organization   *organizationspg.Organization `json:"-" gorm:"foreignKey:OrganizationId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	GroupId        string                        `json:"group_id,omitempty" gorm:"index"`
	// PinHash is the sha256 digest (see library/pinhash) of an 8-digit PIN
	// (see library/random.GeneratePin) -- the plaintext PIN is returned
	// once, on creation, and never stored anywhere.
	PinHash string `json:"-" gorm:"not null;uniqueIndex"`
	// AccountType is "teacher" (default/empty) or "co_director" -- see
	// authorizationpg.MapAccountType for the accepted values. Only a
	// Director (not a CoDirector) may create a co_director invite, enforced
	// in the transport Create handler.
	AccountType string `json:"account_type,omitempty"`
	Name        string `json:"name,omitempty"`
	Email       string `json:"email,omitempty"`
	// PublicDescription is shown to whoever enters the PIN -- safe to
	// return to any authenticated caller looking this invite up.
	PublicDescription string `json:"public_description,omitempty"`
	// InternalNote is visible only to org staff (the create/list
	// endpoints) -- never returned from the lookup endpoint a PIN-enterer
	// calls.
	InternalNote string     `json:"internal_note,omitempty"`
	CreatedBy    string     `json:"created_by,omitempty" gorm:"not null"`
	ExpiresAt    time.Time  `json:"expires_at,omitempty" gorm:"not null"`
	UsedAt       *time.Time `json:"used_at,omitempty"`
	CreateTime   time.Time  `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime   time.Time  `json:"update_time,omitempty" gorm:"not null"`
}

// IsUsable reports whether the invite can still be completed — not
// expired, not already used.
func (i TeacherInvite) IsUsable(now time.Time) bool {
	return i.UsedAt == nil && now.Before(i.ExpiresAt)
}
