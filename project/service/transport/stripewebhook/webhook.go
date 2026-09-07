package stripewebhookgt

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/organizationstripeaccounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/payments"
	"github.com/Johncode248/Rodzic-Chat/variables"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/paymentintent"
	"github.com/stripe/stripe-go/v81/webhook"
)

// Webhook handles POST /public/stripe/webhook -- unauthenticated by design
// (Stripe can't send our JWTs), authenticated instead by verifying the
// Stripe-Signature header against STRIPE_WEBHOOK_SECRET. Configure this URL
// in the Stripe Dashboard once real keys are in place, subscribed to at
// least: checkout.session.completed, account.updated.
//
// Two event types matter here:
//   - checkout.session.completed: a Payment (project/common/payments) was
//     just paid online -- marks it Paid.
//   - account.updated: one organization's Connect account's onboarding/
//     charges/payouts status changed -- kept in sync so
//     GET /private/organizations/stripe/status stays accurate without
//     calling Stripe on every page load.
func Webhook(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Enter(r.Context(), "stripewebhookgt.Webhook")
	var err error
	defer logger.Exit(ctx, "stripewebhookgt.Webhook", &err)

	payload, readErr := io.ReadAll(r.Body)
	if readErr != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// IgnoreAPIVersionMismatch: the Stripe *account*'s default API version
	// (set in the Dashboard, independent of this backend) will often be
	// newer than whatever stripe-go release this backend was built
	// against -- stripe-go's ConstructEvent refuses the event outright on
	// any mismatch by default, which would permanently break this webhook
	// every time Stripe ships a newer default API version, with no code
	// change on our side to trigger it. Safe to ignore here because
	// handleCheckoutCompleted/handleAccountUpdated below never rely on
	// stripe-go's own versioned struct parsing -- they re-`json.Unmarshal`
	// event.Data.Raw into just the handful of stable fields (id, status,
	// charges/payouts/details flags) this backend actually reads.
	event, err := webhook.ConstructEventWithOptions(
		payload, r.Header.Get("Stripe-Signature"), variables.STRIPEWebhook,
		webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true},
	)
	if err != nil {
		http.Error(w, "Invalid signature", http.StatusBadRequest)
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		err = handleCheckoutCompleted(ctx, event)
	case "account.updated":
		err = handleAccountUpdated(ctx, event)
	default:
		// Anything else we didn't subscribe to on purpose -- ack it so
		// Stripe stops retrying, but there's nothing to do.
	}

	if err != nil {
		log.Printf("stripewebhookgt.Webhook: error handling %s: %v", event.Type, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleCheckoutCompleted(ctx context.Context, event stripe.Event) error {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		return err
	}

	if session.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
		// e.g. an async payment method still pending -- Stripe will send a
		// follow-up checkout.session.async_payment_succeeded/failed event
		// we don't currently subscribe to; nothing to record yet.
		return nil
	}

	paymentIntentId := ""
	if session.PaymentIntent != nil {
		paymentIntentId = session.PaymentIntent.ID
	}

	actualStripeFeeGrosze := int64(0)
	if paymentIntentId != "" {
		// Best-effort: the real Stripe fee is purely informational (see
		// Payment's doc comment), so a failure here must never block
		// marking the payment Paid.
		if pi, err := paymentintent.Get(paymentIntentId, &stripe.PaymentIntentParams{
			Expand: []*string{stripe.String("latest_charge.balance_transaction")},
		}); err == nil && pi.LatestCharge != nil && pi.LatestCharge.BalanceTransaction != nil {
			actualStripeFeeGrosze = pi.LatestCharge.BalanceTransaction.Fee
		} else if err != nil {
			log.Printf("stripewebhookgt: could not resolve real Stripe fee for payment_intent %s: %v", paymentIntentId, err)
		}
	}

	return payments.ApplyCheckoutCompleted(ctx, session.ID, paymentIntentId, actualStripeFeeGrosze)
}

func handleAccountUpdated(ctx context.Context, event stripe.Event) error {
	var account stripe.Account
	if err := json.Unmarshal(event.Data.Raw, &account); err != nil {
		return err
	}

	return organizationstripeaccounts.SyncFromStripeAccount(ctx, account.ID, account.ChargesEnabled, account.PayoutsEnabled, account.DetailsSubmitted)
}
