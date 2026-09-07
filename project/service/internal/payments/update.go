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

	if request.Title != nil {
		payment.Title = *request.Title
	}
	if request.Description != nil {
		payment.Description = *request.Description
	}
	if request.BaseAmountGrosze != nil {
		if *request.BaseAmountGrosze <= 0 {
			return res, ErrInvalidAmount
		}
		payment.BaseAmountGrosze = *request.BaseAmountGrosze
	}
	if request.DueDate != nil {
		payment.DueDate = request.DueDate
	}

	if request.Status != nil {
		status := paymentspg.MapPaymentStatus(*request.Status)
		if status == paymentspg.PaymentStatusUnknown {
			return res, ErrUnknownStatus
		}

		now := time.Now()
		payment.Status = status

		switch status {
		case paymentspg.PaymentStatusPaid:
			if payment.PaidAt == nil {
				payment.PaidAt = &now
			}
		case paymentspg.PaymentStatusReconciled:
			payment.ReconciledAt = &now
			payment.ReconciledBy = actingAccountId
			// A payment reconciled straight from Unpaid (no online checkout
			// ever created) was necessarily settled outside the app -- a
			// bank transfer, per product direction, which is why it never
			// picked up an online PaymentMethod along the way.
			if payment.PaymentMethod == paymentspg.PaymentMethodUnset {
				payment.PaymentMethod = paymentspg.PaymentMethodBankTransfer
			}
			if payment.PaidAt == nil {
				payment.PaidAt = &now
			}
		}
	}

	if err = gormdb.Update(ctx, &payment, query); err != nil {
		return res, err
	}

	res.Payment = payment

	return res, nil
}
