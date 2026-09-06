package accounts

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
)

func Select(ctx context.Context, argument accountspg.SelectArgument) (res accountspg.SelectResponse, err error) {
	ctx = logger.Enter(ctx, "accounts.Select", argument)
	defer logger.Exit(ctx, "accounts.Select", &res, &err)

	var (
		usesMutation = &accountspg.SelectMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = ProcessSelectMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return accountspg.SelectResponse{
		Accounts: usesMutation.Transient,
	}, nil
}

func ProcessSelectMutation(ctx context.Context, mutation *accountspg.SelectMutation) (err error) {
	ctx = logger.Enter(ctx, "accounts.ProcessSelectMutation", mutation)
	defer logger.Exit(ctx, "accounts.ProcessSelectMutation", &err)

	result, err := gormdb.Select[accountspg.Account](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = result

	return nil
}
