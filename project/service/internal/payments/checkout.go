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
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
)

var ErrNotYourPayment = errors.New("this payment is not assigned to you")
var ErrPaymentNotPayable = errors.New("this payment is not awaiting payment")
var ErrOrganizationNotReadyForPayments = errors.New("this organization has not finished setting up online payments yet")

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

	successURL := fmt.Sprintf("%s/platnosci/potwierdzenie?session_id={CHECKOUT_SESSION_ID}", variables.DOMEIN_F)
	cancelURL := fmt.Sprintf("%s/platnosci/anulowano?payment_id=%s", variables.DOMEIN_F, payment.Id)

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

	payment.PlatformFeeGrosze = amounts.PlatformFeeGrosze
	payment.StripeFeeEstimateGrosze = amounts.StripeFeeEstimateGrosze
	payment.TotalAmountGrosze = amounts.TotalAmountGrosze
	payment.PaymentMethod = paymentspg.PaymentMethodOnline
	payment.StripeCheckoutSessionId = checkoutSession.ID
	payment.UpdateTime = time.Now()

	if err = gormdb.Update(ctx, &payment, databasepg.Query{
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: payment.Id}}},
		},
	}); err != nil {
		return res, err
	}

	res.Url = checkoutSession.URL
	res.TotalAmountGrosze = amounts.TotalAmountGrosze

	return res, nil
}
