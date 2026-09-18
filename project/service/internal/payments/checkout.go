package payments

import (
	"context"
	"errors"
	"fmt"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	paymentspg "github.com/Johncode248/Rodzic-Chat/project/common/payments"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/organizationstripeaccounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/platformsettings"
	"github.com/Johncode248/Rodzic-Chat/variables"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
)

var ErrNotYourPayment = errors.New("this payment is not assigned to you")
var ErrPaymentNotPayable = errors.New("this payment is not awaiting payment")
var ErrOrganizationNotReadyForPayments = errors.New("this organization has not finished setting up online payments yet")
var ErrCheckoutInProgress = errors.New("a checkout for this payment was just started -- please wait a moment and try again")

// checkoutClaimStaleAfter bounds how long a claimed-but-never-finished
// checkout attempt (e.g. the process crashed between claiming and calling
// Stripe) blocks a retry.
const checkoutClaimStaleAfter = 5 * time.Minute

// CreateCheckoutSession starts a Stripe Checkout Session for one Unpaid
// Payment that belongs to parentId, charging the parent BaseAmount +
// PlatformFee + an estimate of Stripe's own cut (see ComputeCheckoutAmounts)
// so the organization's connected account still receives exactly
// BaseAmountGrosze via a pinned transfer -- see the "destination charge"
// call below.
func CreateCheckoutSession(ctx context.Context, paymentId, organizationId, parentId string) (res paymentspg.CheckoutResponse, err error) {
	ctx = logger.Enter(ctx, "payments.CreateCheckoutSession", paymentId, organizationId, parentId)
	defer logger.Exit(ctx, "payments.CreateCheckoutSession", &res, &err)

	rows, err := gormdb.Select[paymentspg.Payment](ctx, databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{
				Filters: []databasepg.Filter{
					{Key: "id", Operator: "=", Value: paymentId},
					{Key: "organization_id", Operator: "=", Value: organizationId},
				},
			},
		},
	})
	if err != nil {
		return res, err
	}
	if len(rows) == 0 {
		return res, ErrPaymentNotFound
	}
	payment := rows[0]

	if payment.ParentId != parentId {
		return res, ErrNotYourPayment
	}
	if payment.Status != paymentspg.PaymentStatusUnpaid {
		return res, ErrPaymentNotPayable
	}

	// Atomically claim this payment before talking to Stripe: a double
	// click (or a retried request) racing this same function concurrently
	// would otherwise both pass the Unpaid check above, both create their
	// own Stripe Checkout Session, and the second gormdb.Update below would
	// silently overwrite the first session's id -- so if the customer pays
	// via the first (now orphaned) session, ApplyCheckoutCompleted's lookup
	// by stripe_checkout_session_id never finds a match and the payment is
	// never marked Paid even though Stripe captured the money. The UPDATE
	// below only succeeds for one concurrent caller because it's a single
	// atomic statement gated on the row still looking unclaimed (either
	// never attempted, or a stale attempt older than checkoutClaimStaleAfter
	// -- covering a crash between claiming and finishing).
	claimToken := "pending:" + uuid.NewString()
	claimCutoff := time.Now().Add(-checkoutClaimStaleAfter)
	claim := gormdb.GormDB.WithContext(ctx).Model(&paymentspg.Payment{}).
		Where("id = ? AND status = ? AND (stripe_checkout_session_id = '' OR update_time < ?)",
			payment.Id, paymentspg.PaymentStatusUnpaid, claimCutoff).
		Updates(map[string]interface{}{
			"stripe_checkout_session_id": claimToken,
			"update_time":                time.Now(),
		})
	if claim.Error != nil {
		return res, claim.Error
	}
	if claim.RowsAffected == 0 {
		return res, ErrCheckoutInProgress
	}
	// If anything below fails, release the claim so the caller can retry
	// immediately instead of waiting out checkoutClaimStaleAfter.
	succeeded := false
	defer func() {
		if succeeded {
			return
		}
		gormdb.GormDB.WithContext(ctx).Model(&paymentspg.Payment{}).
			Where("id = ? AND stripe_checkout_session_id = ?", payment.Id, claimToken).
			Update("stripe_checkout_session_id", "")
	}()

	stripeAccount, err := organizationstripeaccounts.Get(ctx, organizationId)
	if err != nil {
		if errors.Is(err, organizationstripeaccounts.ErrNoAccountYet) {
			return res, ErrOrganizationNotReadyForPayments
		}
		return res, err
	}
	if !stripeAccount.ChargesEnabled {
		return res, ErrOrganizationNotReadyForPayments
	}

	settings, err := platformsettings.Get(ctx)
	if err != nil {
		return res, err
	}

	amounts := ComputeCheckoutAmounts(payment.BaseAmountGrosze, settings)

	successURL := fmt.Sprintf("%s/platnosci/potwierdzenie?session_id={CHECKOUT_SESSION_ID}", variables.FrontendURL)
	cancelURL := fmt.Sprintf("%s/platnosci/anulowano?payment_id=%s", variables.FrontendURL, payment.Id)

	productData := &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
		Name: stripe.String(payment.Title),
	}
	// Stripe rejects an explicit empty string here ("cannot be unset") --
	// Description must be omitted entirely rather than sent as "" when a
	// payment has no description (the common case: it's an optional field
	// on Payment, see project/common/payments/payment.go).
	if payment.Description != "" {
		productData.Description = stripe.String(payment.Description)
	}

	checkoutSession, err := session.New(&stripe.CheckoutSessionParams{
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		// BLIK only: its Stripe fee (1.6% + 1.00 zl) is fixed and known up front,
		// unlike "card" whose real cost varies by issuing bank/country -- see
		// ComputeCheckoutAmounts's doc comment. Restricting to just this one
		// method also excludes Card/Apple Pay/Google Pay/Klarna/Link, which
		// Stripe would otherwise offer alongside it by default.
		PaymentMethodTypes: []*string{
			stripe.String("blik"),
		},
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Quantity: stripe.Int64(1),
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency:    stripe.String("pln"),
					UnitAmount:  stripe.Int64(amounts.TotalAmountGrosze),
					ProductData: productData,
				},
			},
		},
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			TransferData: &stripe.CheckoutSessionPaymentIntentDataTransferDataParams{
				Destination: stripe.String(stripeAccount.StripeAccountId),
				Amount:      stripe.Int64(payment.BaseAmountGrosze),
			},
			Metadata: map[string]string{
				"payment_id":      payment.Id,
				"organization_id": organizationId,
			},
		},
		Metadata: map[string]string{
			"payment_id":      payment.Id,
			"organization_id": organizationId,
		},
	})
	if err != nil {
		return res, err
	}

	// A narrow, column-scoped update, not gormdb.Update(ctx, &payment, ...)
	// -- that helper does Select("*").Updates(model), a full-row overwrite
	// using the `payment` struct as it was read at the top of this
	// function, BEFORE the Stripe network round-trip above. If staff PATCH
	// this same payment while that round-trip is in flight (e.g. marking it
	// Reconciled for a bank transfer received in parallel), a full-row
	// overwrite here would silently revert that change back to the stale
	// pre-checkout values, including Status: Unpaid, while a live Stripe
	// session is now attached. Only writing the columns this function
	// actually computed leaves any concurrent change to other columns
	// (Status included) intact.
	if err = gormdb.GormDB.WithContext(ctx).Model(&paymentspg.Payment{}).
		Where("id = ?", payment.Id).
		Updates(map[string]interface{}{
			"platform_fee_grosze":        amounts.PlatformFeeGrosze,
			"stripe_fee_estimate_grosze": amounts.StripeFeeEstimateGrosze,
			"total_amount_grosze":        amounts.TotalAmountGrosze,
			"payment_method":             paymentspg.PaymentMethodOnline,
			"stripe_checkout_session_id": checkoutSession.ID,
			"update_time":                time.Now(),
		}).Error; err != nil {
		// The Stripe Checkout Session already exists at this point --
		// session.New succeeded above. If this write fails, best-effort
		// persist JUST the real session id (ignoring any further error)
		// instead of letting the deferred release below reset the column
		// back to "". Without this, the row keeps claimToken, the defer's
		// conditional WHERE matches it and clears it to "", and the
		// session becomes orphaned: a customer can still complete payment
		// against it at Stripe, but ApplyCheckoutCompleted's webhook
		// lookup by stripe_checkout_session_id will never find this
		// payment, so it's never marked Paid despite Stripe having
		// captured the money. Writing the real id here means that lookup
		// still succeeds even though the fee/method columns didn't get
		// updated (they'll be stale until reconciled some other way, but
		// the payment itself won't silently swallow captured money).
		gormdb.GormDB.WithContext(ctx).Model(&paymentspg.Payment{}).
			Where("id = ? AND stripe_checkout_session_id = ?", payment.Id, claimToken).
			Update("stripe_checkout_session_id", checkoutSession.ID)
		return res, err
	}
	succeeded = true

	res.Url = checkoutSession.URL
	res.TotalAmountGrosze = amounts.TotalAmountGrosze

	return res, nil
}
