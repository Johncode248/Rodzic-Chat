package payments

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	paymentspg "github.com/Johncode248/Rodzic-Chat/project/common/payments"
)

func Select(ctx context.Context, query databasepg.Query) (res paymentspg.SelectResponse, err error) {
	ctx = logger.Enter(ctx, "payments.Select", query)
	defer logger.Exit(ctx, "payments.Select", &res, &err)

	rows, err := gormdb.Select[paymentspg.Payment](ctx, query)
	if err != nil {
		return res, err
	}

	res.Payments = rows

	return res, nil
}
