package caregiverspg

import "time"

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
	Id          string        `json:"id,omitempty" gorm:"primaryKey"`
	AccountId   string        `json:"account_id,omitempty" gorm:"not null;index;uniqueIndex:idx_caregiver_account_type"`
	Type        CaregiverType `json:"type,omitempty" gorm:"not null;uniqueIndex:idx_caregiver_account_type"`
	DisplayName string        `json:"display_name,omitempty" gorm:"not null"`
	Description string        `json:"description,omitempty"`
	Experience  string        `json:"experience,omitempty"`
	// HourlyRate is in grosze (see this repo's money convention) -- 0 means
	// not given, not "free".
	HourlyRate int64     `json:"hourly_rate,omitempty"`
	City       string    `json:"city,omitempty"`
	Phone      string    `json:"phone,omitempty"`
	Email      string    `json:"email,omitempty"`
	CreateTime time.Time `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime time.Time `json:"update_time,omitempty" gorm:"not null"`
}
