package groupspg

import (
	"time"
)

type Group struct {
	Id             string `json:"id,omitempty" gorm:"primaryKey"`
	OrganizationId string `json:"organization_id,omitempty" gorm:"not null;index"`
	Name           string `json:"name,omitempty" gorm:"not null"`
	TeacherId      string `json:"teacher_id,omitempty"`
	AgeRange       string `json:"age_range,omitempty"`
	// RequireChildBirthDateVerification: when true, a parent registering
	// through this group's registration link must provide the child's
	// birth date (whether creating a new child or matching an existing
	// one by name). Matching against an existing child then requires the
	// date to agree with what's on file -- a mismatch (or a second parent
	// registering when the flag is off and this wasn't collected) creates
	// a GuardianClaim for the director to approve/reject instead of
	// auto-linking, closing the "guess a classmate's name, get pickup
	// rights" hole that pure name-matching has on its own. Director-toggled,
	// set at group creation (see project/common/guardianclaims).
	RequireChildBirthDateVerification bool      `json:"require_child_birth_date_verification,omitempty"`
	CreateTime                        time.Time `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime                        time.Time `json:"update_time,omitempty" gorm:"not null"`
}
