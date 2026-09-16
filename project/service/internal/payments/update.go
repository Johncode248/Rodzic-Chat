package payments

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	paymentspg "github.com/Johncode248/Rodzic-Chat/project/common/payments"
)

var ErrPaymentNotFound = errors.New("payment not found")
var ErrCannotEditNonUnpaidPayment = errors.New("title/description/amount/due date can only be changed while the payment is still unpaid")
var ErrUnknownStatus = errors.New("unrecognized status")

// Update applies a partial change to one Payment, scoped to organizationId
// (so staff can never touch another org's payment by guessing an id).
// Title/Description/BaseAmountGrosze/DueDate are rejected once the payment
// has moved off Unpaid (see Payment's doc comment) -- Status itself can
// always be changed, since that's how a director marks a bank-transfer
// payment Reconciled, or corrects a stuck online payment by hand.
func Update(ctx context.Context, id, organizationId, actingAccountId string, request paymentspg.UpdateRequest) (res paymentspg.UpdateResponse, err error) {
	ctx = logger.Enter(ctx, "payments.Update", id, organizationId, actingAccountId, request)
	defer logger.Exit(ctx, "payments.Update", &res, &err)

	query := databasepg.Query{
		Groups: []databasepg.Group{
			{
				Filters: []databasepg.Filter{
					{Key: "id", Operator: "=", Value: id},
					{Key: "organization_id", Operator: "=", Value: organizationId},
				},
			},
		},
	}

	rows, err := gormdb.Select[paymentspg.Payment](ctx, databasepg.Query{Limit: 1, Groups: query.Groups})
	if err != nil {
		return res, err
	}
	if len(rows) == 0 {
		return res, ErrPaymentNotFound
	}
	payment := rows[0]

	editingFields := request.Title != nil || request.Description != nil ||
		request.BaseAmountGrosze != nil || request.DueDate != nil
	if editingFields && payment.Status != paymentspg.PaymentStatusUnpaid {
		return res, ErrCannotEditNonUnpaidPayment
	}

	// Collected into a narrow, column-scoped update below (not
	// gormdb.Update(ctx, &payment, query), which does a full-row
	// Select("*").Updates(model) overwrite using this in-memory `payment`
	// struct) -- a full-row write here would silently clobber any column
	// a CONCURRENT write touched that this request didn't intend to
	// change, e.g. CreateCheckoutSession attaching a live
	// stripe_checkout_session_id/payment_method while this staff edit was
	// in flight, or ApplyCheckoutCompleted marking the payment Paid. Only
	// writing the columns this function actually computed leaves any such
	// concurrent change to other columns intact. Mirrors the same fix
	// already applied to payments.CreateCheckoutSession.
	changes := map[string]interface{}{}

	if request.Title != nil {
		payment.Title = *request.Title
		changes["title"] = payment.Title
	}
	if request.Description != nil {
		payment.Description = *request.Description
		changes["description"] = payment.Description
	}
	if request.BaseAmountGrosze != nil {
		if *request.BaseAmountGrosze <= 0 {
			return res, ErrInvalidAmount
		}
		payment.BaseAmountGrosze = *request.BaseAmountGrosze
		changes["base_amount_grosze"] = payment.BaseAmountGrosze
	}
	if request.DueDate != nil {
		payment.DueDate = request.DueDate
		changes["due_date"] = payment.DueDate
	}

	if request.Status != nil {
		status := paymentspg.MapPaymentStatus(*request.Status)
		if status == paymentspg.PaymentStatusUnknown {
			return res, ErrUnknownStatus
		}

		now := time.Now()
		payment.Status = status
		changes["status"] = payment.Status

		switch status {
		case paymentspg.PaymentStatusPaid:
			if payment.PaidAt == nil {
				payment.PaidAt = &now
				changes["paid_at"] = payment.PaidAt
			}
		case paymentspg.PaymentStatusReconciled:
			payment.ReconciledAt = &now
			payment.ReconciledBy = actingAccountId
			changes["reconciled_at"] = payment.ReconciledAt
			changes["reconciled_by"] = payment.ReconciledBy
			// A payment reconciled straight from Unpaid (no online checkout
			// ever created) was necessarily settled outside the app -- a
			// bank transfer, per product direction, which is why it never
			// picked up an online PaymentMethod along the way.
			if payment.PaymentMethod == paymentspg.PaymentMethodUnset {
				payment.PaymentMethod = paymentspg.PaymentMethodBankTransfer
				changes["payment_method"] = payment.PaymentMethod
			}
			if payment.PaidAt == nil {
				payment.PaidAt = &now
				changes["paid_at"] = payment.PaidAt
			}
		}
	}

	if len(changes) > 0 {
		changes["update_time"] = time.Now()
		if err = gormdb.GormDB.WithContext(ctx).Model(&paymentspg.Payment{}).
			Where("id = ? AND organization_id = ?", id, organizationId).
			Updates(changes).Error; err != nil {
			return res, err
		}
	}

	res.Payment = payment

	return res, nil
}
