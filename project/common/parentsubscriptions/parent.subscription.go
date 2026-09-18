package parentsubscriptionspg

import (
	"time"

	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
)

// ParentSubscription is one parent account's plan tier -- MVP only, so it holds nothing but
// which plan the account is on. One row per Account (unique AccountId), created lazily on first
// read/use, defaulting to PlanFree (see project/service/internal/parentsubscriptions.GetOrCreate).
// Usage/quota tracking (AI tokens, Google Places searches, etc.) and Stripe fields were
// deliberately stripped out here -- they weren't enforcing anything real yet, and the one feature
// that did use them (the AI assistant) is hard-disabled server-side for now (see
// project/service/transport/aichat.Send). Re-add that bookkeeping alongside whichever feature
// actually needs it again.
//
// TEST-MODE NOTE (see transport/parentsubscriptions/test.set.plan.go): real billing for this is
// required by Apple/Google to go through StoreKit/Play Billing (in-app purchase), not Stripe --
// this system currently only supports a manual test-plan switch so the app's UI can be built and
// verified before real IAP is wired up.
type ParentSubscription struct {
	Id         string              `json:"id,omitempty" gorm:"primaryKey"`
	AccountId  string              `json:"account_id,omitempty" gorm:"unique;not null"`
	Account    *accountspg.Account `json:"-" gorm:"foreignKey:AccountId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Plan       Plan                `json:"plan,omitempty" gorm:"not null"`
	CreateTime time.Time           `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime time.Time           `json:"update_time,omitempty" gorm:"not null"`
}
