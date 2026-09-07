package payments

import (
	"context"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	paymentspg "github.com/Johncode248/Rodzic-Chat/project/common/payments"
)

// ApplyCheckoutCompleted marks the Payment matching stripeCheckoutSessionId
// as Paid -- called from the Stripe webhook on checkout.session.completed.
// A no-op (not an error) if no matching row is found, since a stray/replayed
// webhook event shouldn't crash the handler; actualStripeFeeGrosze may be
// zero if we couldn't resolve the real balance transaction (best-effort,
// see transport/stripewebhook -- it's purely informational either way, see
// Payment's doc comment on why it never affects what anyone owes/receives).
func ApplyCheckoutCompleted(ctx context.Context, stripeCheckoutSessionId, stripePaymentIntentId string, actualStripeFeeGrosze int64) (err error) {
	ctx = logger.Enter(ctx, "payments.ApplyCheckoutCompleted", stripeCheckoutSessionId, stripePaymentIntentId, actualStripeFeeGrosze)
	defer logger.Exit(ctx, "payments.ApplyCheckoutCompleted", &err)

	query := databasepg.Query{
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "stripe_checkout_session_id", Operator: "=", Value: stripeCheckoutSessionId}}},
		},
	}

	rows, err := gormdb.Select[paymentspg.Payment](ctx, databasepg.Query{Limit: 1, Groups: query.Groups})
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	payment := rows[0]

	// Idempotency: Stripe may deliver the same event more than once.
	if payment.Status == paymentspg.PaymentStatusPaid || payment.Status == paymentspg.PaymentStatusReconciled {
		return nil
	}

	now := time.Now()
	payment.Status = paymentspg.PaymentStatusPaid
	payment.PaidAt = &now
	payment.StripePaymentIntentId = stripePaymentIntentId
	if actualStripeFeeGrosze > 0 {
		payment.ActualStripeFeeGrosze = actualStripeFeeGrosze
	}
	payment.UpdateTime = now

	return gormdb.Update(ctx, &payment, query)
}
