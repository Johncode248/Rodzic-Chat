package caregiverspg

import (
	"time"

	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
)

// Caregiver is a professional listing owned by the account that provides
// the service -- a parent (or anyone) who is themselves a nanny or tutor
// creates/edits exactly this row for their own account (see the unique
// index below), rather than any user being able to add a listing for
// someone else. Other parents browse these read-only and can endorse one
// with a CaregiverRecommendation ("Polecam"). Distinct from the generic
// project/common/places.Place model (which covers arbitrary third-party
// places sourced from Google/OSM/manual entry, no owning account) -- this
// is specifically the self-service "I offer this" directory.
type Caregiver struct {
	Id          string              `json:"id,omitempty" gorm:"primaryKey"`
	AccountId   string              `json:"account_id,omitempty" gorm:"not null;index;uniqueIndex:idx_caregiver_account_type"`
	Account     *accountspg.Account `json:"-" gorm:"foreignKey:AccountId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Type        CaregiverType       `json:"type,omitempty" gorm:"not null;uniqueIndex:idx_caregiver_account_type"`
	DisplayName string              `json:"display_name,omitempty" gorm:"not null"`
	Description string              `json:"description,omitempty"`
	Experience  string              `json:"experience,omitempty"`
	// HourlyRate is in grosze (see this repo's money convention) -- 0 means
	// not given, not "free".
	HourlyRate int64  `json:"hourly_rate,omitempty"`
	City       string `json:"city,omitempty"`
	Phone      string `json:"phone,omitempty"`
	Email      string `json:"email,omitempty"`
	// Subjects is a comma-separated list of what is offered: school subjects for a tutor
	// ("matematyka, angielski"), age groups / duties for a nanny ("0-3 lata, noce").
	Subjects string `json:"subjects,omitempty"`
	// Where the service happens -- any combination. Online needs no area; Stationary means the
	// client comes to the caregiver (City); Travel means the caregiver comes to the client, up to
	// TravelRadiusKm from City / inside ServiceCities.
	Online         bool `json:"online"`
	Stationary     bool `json:"stationary"`
	Travel         bool `json:"travel"`
	TravelRadiusKm int  `json:"travel_radius_km,omitempty"`
	// ServiceCities is a comma-separated list of the cities served (in addition to City).
	ServiceCities string `json:"service_cities,omitempty"`
	// Verification: every new or edited profile waits for a platform admin before it appears in
	// the directory (see ReviewStatus*). RejectReason is shown to the owner.
	ReviewStatus ReviewStatus `json:"review_status" gorm:"not null;default:2;index"`
	RejectReason string       `json:"reject_reason,omitempty"`
	ReviewedAt   *time.Time   `json:"reviewed_at,omitempty"`
	CreateTime   time.Time    `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime   time.Time    `json:"update_time,omitempty" gorm:"not null"`
}

// ReviewStatus of a caregiver profile. New values only ever appended (stored as an int).
// The DB default is Approved so rows created before verification existed stay visible.
type ReviewStatus int

const (
	ReviewStatusUnknown  ReviewStatus = iota
	ReviewStatusPending               // 1: waiting for a platform admin
	ReviewStatusApproved              // 2: visible in the directory
	ReviewStatusRejected              // 3: not visible; owner can fix and resubmit
)
