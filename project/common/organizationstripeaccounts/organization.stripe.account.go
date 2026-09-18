package organizationstripeaccountspg

import (
	"time"

	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
)

// OrganizationStripeAccount links an Organization to its Stripe Connect
// Express account (see project/service/internal/organizationstripeaccounts
// for the onboarding flow). One per organization -- a Payment's online
// checkout (project/common/payments) transfers to whatever account is
// linked here.
//
// ChargesEnabled/PayoutsEnabled/DetailsSubmitted mirror the same-named
// fields on a Stripe Account object, kept in sync via the account.updated
// webhook -- they're what the UI checks to show "Stripe: wymaga
// uzupełnienia danych" vs "połączono" without calling Stripe on every page
// load.
type OrganizationStripeAccount struct {
	Id               string                        `json:"id,omitempty" gorm:"primaryKey"`
	OrganizationId   string                        `json:"organization_id,omitempty" gorm:"unique;not null;index"`
	Organization     *organizationspg.Organization `json:"-" gorm:"foreignKey:OrganizationId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	StripeAccountId  string                        `json:"stripe_account_id,omitempty" gorm:"unique;not null"`
	ChargesEnabled   bool                          `json:"charges_enabled,omitempty"`
	PayoutsEnabled   bool                          `json:"payouts_enabled,omitempty"`
	DetailsSubmitted bool                          `json:"details_submitted,omitempty"`
	CreateTime       time.Time                     `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime       time.Time                     `json:"update_time,omitempty" gorm:"not null"`
}

type SelectResponse struct {
	OrganizationStripeAccounts []OrganizationStripeAccount `json:"organization_stripe_accounts,omitempty"`
}

// OnboardingLinkResponse is what POST /private/organizations/stripe/onboarding-link
// returns: a one-time-use, short-lived URL to Stripe's own hosted
// onboarding/KYC flow for this organization's connected account.
type OnboardingLinkResponse struct {
	Url string `json:"url,omitempty"`
}

// StatusResponse is what GET /private/organizations/stripe/status returns
// -- Connected=false means no Stripe account has been created for this org
// yet (onboarding never started).
type StatusResponse struct {
	Connected        bool `json:"connected"`
	ChargesEnabled   bool `json:"charges_enabled"`
	PayoutsEnabled   bool `json:"payouts_enabled"`
	DetailsSubmitted bool `json:"details_submitted"`
}
