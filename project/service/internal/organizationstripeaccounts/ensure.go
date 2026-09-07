package organizationstripeaccounts

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	organizationstripeaccountspg "github.com/Johncode248/Rodzic-Chat/project/common/organizationstripeaccounts"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/account"
	"github.com/stripe/stripe-go/v81/accountlink"
)

var ErrNoAccountYet = errors.New("this organization has not started Stripe onboarding yet")

// Get returns the OrganizationStripeAccount row for organizationId, if any.
func Get(ctx context.Context, organizationId string) (res organizationstripeaccountspg.OrganizationStripeAccount, err error) {
	ctx = logger.Enter(ctx, "organizationstripeaccounts.Get", organizationId)
	defer logger.Exit(ctx, "organizationstripeaccounts.Get", &res, &err)

	rows, err := gormdb.Select[organizationstripeaccountspg.OrganizationStripeAccount](ctx, databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "organization_id", Operator: "=", Value: organizationId}}},
		},
	})
	if err != nil {
		return res, err
	}
	if len(rows) == 0 {
		return res, ErrNoAccountYet
	}

	return rows[0], nil
}

// EnsureOnboardingLink returns a fresh Stripe Account Link the director can
// open to complete (or resume) Connect Express onboarding -- creating the
// underlying Stripe Account (and its OrganizationStripeAccount row) first
// if this organization has never started the process. Safe to call
// repeatedly: an already-onboarded organization just gets a fresh
// "account_update" link (useful if Stripe later asks for more info), and a
// partially-onboarded one gets a link that resumes where they left off.
func EnsureOnboardingLink(ctx context.Context, organizationId, returnURL, refreshURL string) (res organizationstripeaccountspg.OnboardingLinkResponse, err error) {
	ctx = logger.Enter(ctx, "organizationstripeaccounts.EnsureOnboardingLink", organizationId)
	defer logger.Exit(ctx, "organizationstripeaccounts.EnsureOnboardingLink", &res, &err)

	existing, err := Get(ctx, organizationId)
	linkType := "account_onboarding"
	stripeAccountId := ""

	if err != nil {
		if !errors.Is(err, ErrNoAccountYet) {
			return res, err
		}

		created, createErr := account.New(&stripe.AccountParams{
			Type:    stripe.String(string(stripe.AccountTypeExpress)),
			Country: stripe.String("PL"),
			Capabilities: &stripe.AccountCapabilitiesParams{
				CardPayments: &stripe.AccountCapabilitiesCardPaymentsParams{Requested: stripe.Bool(true)},
				Transfers:    &stripe.AccountCapabilitiesTransfersParams{Requested: stripe.Bool(true)},
			},
		})
		if createErr != nil {
			return res, createErr
		}

		row := organizationstripeaccountspg.OrganizationStripeAccount{
			Id:              uuid.NewString(),
			OrganizationId:  organizationId,
			StripeAccountId: created.ID,
			CreateTime:      time.Now(),
			UpdateTime:      time.Now(),
		}
		if err = gormdb.Create(ctx, &row); err != nil {
			return res, err
		}

		stripeAccountId = created.ID
	} else {
		stripeAccountId = existing.StripeAccountId
		if !existing.DetailsSubmitted {
			linkType = "account_onboarding"
		} else {
			linkType = "account_update"
		}
	}

	link, err := accountlink.New(&stripe.AccountLinkParams{
		Account:    stripe.String(stripeAccountId),
		ReturnURL:  stripe.String(returnURL),
		RefreshURL: stripe.String(refreshURL),
		Type:       stripe.String(linkType),
	})
	if err != nil {
		return res, err
	}

	res.Url = link.URL

	return res, nil
}

// SyncFromStripeAccount applies an account.updated webhook payload's status
// flags to the matching OrganizationStripeAccount row. A no-op (not an
// error) if we don't recognize the Stripe account id -- that happens for
// any Stripe account that isn't one of ours (shouldn't occur given our
// webhook only subscribes to our own Connect accounts, but a stray event
// shouldn't crash the handler).
func SyncFromStripeAccount(ctx context.Context, stripeAccountId string, chargesEnabled, payoutsEnabled, detailsSubmitted bool) (err error) {
	ctx = logger.Enter(ctx, "organizationstripeaccounts.SyncFromStripeAccount", stripeAccountId, chargesEnabled, payoutsEnabled, detailsSubmitted)
	defer logger.Exit(ctx, "organizationstripeaccounts.SyncFromStripeAccount", &err)

	rows, err := gormdb.Select[organizationstripeaccountspg.OrganizationStripeAccount](ctx, databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "stripe_account_id", Operator: "=", Value: stripeAccountId}}},
		},
	})
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	row := rows[0]
	row.ChargesEnabled = chargesEnabled
	row.PayoutsEnabled = payoutsEnabled
	row.DetailsSubmitted = detailsSubmitted

	return gormdb.Update(ctx, &row, databasepg.Query{
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: row.Id}}},
		},
	})
}
