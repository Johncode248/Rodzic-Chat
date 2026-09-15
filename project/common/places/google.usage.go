package placespg

import "time"

// GooglePlacesUsage is a single global counter row per calendar month (PeriodMonth = "2026-09"
// etc.) -- tracks how many Google Places API calls THIS WHOLE APP has made this month, across
// every parent, so the backend can stay under Google's free monthly quota (see
// MonthlyGooglePlacesRequestBudget) regardless of how many parents are searching. Distinct from
// the PER-PARENT daily cap (parentsubscriptionspg.DailyGooglePlacesSearchLimit), which throttles
// one parent from burning through the shared budget alone -- both must have room for a search to
// actually use Google Places (see project/service/internal/places' search orchestration); either
// one being exhausted falls back to OpenStreetMap instead of erroring.
type GooglePlacesUsage struct {
	Id           string    `json:"id,omitempty" gorm:"primaryKey"`
	PeriodMonth  string    `json:"period_month,omitempty" gorm:"uniqueIndex;not null"`
	RequestsUsed int64     `json:"requests_used,omitempty" gorm:"not null;default:0"`
	CreateTime   time.Time `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime   time.Time `json:"update_time,omitempty" gorm:"not null"`
}

// MonthlyGooglePlacesRequestBudget is kept comfortably under Google's own free-tier monthly
// quota (5,000 requests) so a slightly-stale read-then-write counter increment (see
// project/service/internal/places, no DB-level atomic increment used here) can never actually
// tip the account into a billed request even under a burst of concurrent searches.
const MonthlyGooglePlacesRequestBudget = 4_800
