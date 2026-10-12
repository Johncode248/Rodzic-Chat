package childrenpg

import (
	"time"
)

// Child.OrganizationId is empty ("") for a "personal" child a parent added
// entirely on their own (see project/service/internal/children/mine.go),
// with no placowka involved at all -- distinct from a placowka-connected
// child, which always has both OrganizationId and (usually) GroupId set.
// A personal child never appears in any staff-facing roster, never counts
// against an organization's ChildrenLimit, and can optionally be linked to
// a placowka later through the ordinary registration-link flow.
type Child struct {
	Id             string    `json:"id,omitempty" gorm:"primaryKey"`
	OrganizationId string    `json:"organization_id,omitempty" gorm:"index"`
	GroupId        string    `json:"group_id,omitempty" gorm:"index"`
	FirstName      string    `json:"first_name,omitempty" gorm:"not null"`
	LastName       string    `json:"last_name,omitempty" gorm:"not null"`
	BirthDate      time.Time `json:"birth_date,omitempty"`
	Notes          string    `json:"notes,omitempty"`
	// EncryptedPesel is AES-256-GCM ciphertext (see library/crypto), never
	// serialized directly -- it never appears in the ordinary
	// SelectMine/list responses at all, only HasPesel (computed at the
	// transport layer) does. A parent must explicitly call
	// GET /private/children/mine/pesel to decrypt and see the real value
	// (the "eye icon" flow), which is also the only place this ever gets
	// decrypted.
	EncryptedPesel string `json:"-"`
	// BloodType and HealthNote are the child's health card header (see project/common/childhealth), stored on the child
	// itself because the card is strictly 1:1 with it. Never serialized with the child: only the health endpoints, which
	// are limited to the child's linked parents, expose them.
	BloodType  string    `json:"-"`
	HealthNote string    `json:"-"`
	CreateTime time.Time `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime time.Time `json:"update_time,omitempty" gorm:"not null"`
	// HasPesel is computed (never stored) so clients can tell "no PESEL on file yet" apart from
	// "there is one, still masked" without decrypting anything -- set by whichever internal
	// function last touched EncryptedPesel (see ProcessSelectMutation, children.UpdateMine).
	HasPesel bool `json:"has_pesel" gorm:"-"`
}
