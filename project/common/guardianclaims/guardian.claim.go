package guardianclaimspg

import "time"

// GuardianClaim exists because pure name-matching (see
// findOrCreateChild in registrationlinks) lets anyone holding a group's
// shared registration link claim to be a second parent/guardian of any
// child in that group just by typing their name correctly -- a real risk
// since the link is shared across an entire group, not private per family.
//
// When a group has RequireChildBirthDateVerification enabled
// (project/common/groups), a second registrant matching an existing child
// must also provide that child's birth date; if it doesn't match what's on
// file, the registration does NOT auto-link -- it creates one of these
// rows instead, and a director must explicitly approve or reject it before
// any ParentChild row (and the pickup rights that come with it) is created.
type GuardianClaim struct {
	Id             string `json:"id,omitempty" gorm:"primaryKey"`
	OrganizationId string `json:"organization_id,omitempty" gorm:"not null;index"`
	ChildId        string `json:"child_id,omitempty" gorm:"not null;index"`
	ClaimantId     string `json:"claimant_id,omitempty" gorm:"not null"`
	// ClaimantName/Email are a snapshot of what the claimant submitted at
	// registration time -- kept alongside the account row so the director
	// review screen doesn't need extra joins and still makes sense even if
	// the account is later edited.
	ClaimantName  string `json:"claimant_name,omitempty" gorm:"not null"`
	ClaimantEmail string `json:"claimant_email,omitempty" gorm:"not null"`
	// ClaimantBirthDate is what the claimant typed for the child;
	// ChildBirthDateOnFile is a snapshot of what was already stored on the
	// Child row at the moment of the claim (i.e. what the first parent
	// provided) -- shown side by side so the director can see exactly
	// where the two submissions disagree.
	ClaimantBirthDate    time.Time           `json:"claimant_birth_date,omitempty" gorm:"not null"`
	ChildBirthDateOnFile time.Time           `json:"child_birth_date_on_file,omitempty" gorm:"not null"`
	Status               GuardianClaimStatus `json:"status,omitempty" gorm:"not null"`
	CreateTime           time.Time           `json:"create_time,omitempty" gorm:"not null"`
	ResolvedAt           *time.Time          `json:"resolved_at,omitempty"`
	ResolvedBy           string              `json:"resolved_by,omitempty"`
}
