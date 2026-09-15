package parentsubscriptionspg

import "time"

// ParentSubscription is one parent account's AI plan + usage counter for
// the current period. One row per Account (unique AccountId) -- created
// lazily on first read/use, defaulting to PlanFree (see
// project/service/internal/parentsubscriptions.GetOrCreate).
//
// TEST-MODE NOTE (see transport/parentsubscriptions/test.set.plan.go): real
// billing for this is required by Apple/Google to go through StoreKit/Play
// Billing (in-app purchase), not Stripe -- this system currently only
// supports a manual test-plan switch so the app's UI/limits can be built
// and verified before real IAP is wired up. StripeCustomerId/
// StripeSubscriptionId are reserved for a possible future web-based
// upgrade path (Stripe can't be the *only* purchase path once mobile ships
// per store policy, but nothing stops it existing as an option on web).
type ParentSubscription struct {
	Id             string `json:"id,omitempty" gorm:"primaryKey"`
	AccountId      string `json:"account_id,omitempty" gorm:"unique;not null"`
	Plan           Plan   `json:"plan,omitempty" gorm:"not null"`
	AiRequestsUsed int64  `json:"ai_requests_used,omitempty" gorm:"not null;default:0"`
	// AiTokensUsed is the actual enforcement metric for the AI assistant's usage cap (see
	// MonthlyAiTokenLimit) -- a tool-use conversation's real cost varies a lot by how many
	// NFZ/OpenStreetMap round trips it takes, so counting tokens is a much tighter cost control
	// than counting requests. AiRequestsUsed is still tracked alongside it purely for display.
	AiTokensUsed  int64     `json:"ai_tokens_used,omitempty" gorm:"not null;default:0"`
	PeriodResetAt time.Time `json:"period_reset_at,omitempty" gorm:"not null"`
	// GooglePlacesUsedToday/GooglePlacesUsageDate cap how many Google Places searches this one
	// parent can trigger per day (see project/service/internal/places' search orchestration) --
	// resets whenever GooglePlacesUsageDate is no longer "today", the same lazy-rollover pattern
	// PeriodResetAt uses for the monthly AI counters.
	GooglePlacesUsedToday int64     `json:"google_places_used_today,omitempty" gorm:"not null;default:0"`
	GooglePlacesUsageDate time.Time `json:"-"`
	StripeCustomerId      string    `json:"-"`
	StripeSubscriptionId  string    `json:"-"`
	CreateTime            time.Time `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime            time.Time `json:"update_time,omitempty" gorm:"not null"`
}
