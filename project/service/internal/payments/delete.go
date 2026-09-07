package payments

import (
	"context"
	"errors"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	paymentspg "github.com/Johncode248/Rodzic-Chat/project/common/payments"
)

var ErrCannotDeleteNonUnpaidPayment = errors.New("only an unpaid payment can be deleted")

// Delete removes one Payment, scoped to organizationId, but only while it's
// still Unpaid -- once money has moved (or been recorded as moved), the row
// stays as an audit trail; staff can still cancel it going forward by other
// means (e.g. this will likely grow a Cancelled status before v1 ships),
// but silent deletion of a paid/reconciled record isn't safe to allow.
func Delete(ctx context.Context, id, organizationId string) (res paymentspg.DeleteResponse, err error) {
	ctx = logger.Enter(ctx, "payments.Delete", id, organizationId)
	defer logger.Exit(ctx, "payments.Delete", &res, &err)

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
	if rows[0].Status != paymentspg.PaymentStatusUnpaid {
		return res, ErrCannotDeleteNonUnpaidPayment
	}

	// gormdb.Delete's generic Delete(&result) doesn't populate result with
	// the deleted rows' content (it's a plain DELETE, not a RETURNING
	// fetch) -- rows[0], already SELECTed above, is what the caller
	// actually wants to see confirmed-deleted.
	if _, err = gormdb.Delete[paymentspg.Payment](ctx, query); err != nil {
		return res, err
	}

	res.Payments = rows

	return res, nil
}
