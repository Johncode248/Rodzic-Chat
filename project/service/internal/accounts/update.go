package accounts

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
)

func Update(ctx context.Context, argument accountspg.UpdateArgument) (res accountspg.UpdateResponse, err error) {
	ctx = logger.Enter(ctx, "accounts.Update", argument)
	defer logger.Exit(ctx, "accounts.Update", &res, &err)

	var (
		usesMutation = &accountspg.UpdateMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = InitiateUpdateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	defer func() {
		if err != nil {
			if err = RollbackUpdateMutation(ctx, usesMutation); err != nil {
				panic(err)
			}
		}
	}()

	if err = FinalizeUpdateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return accountspg.UpdateResponse{
		Account: usesMutation.Transient.Current,
	}, nil
}

func InitiateUpdateMutation(ctx context.Context, mutation *accountspg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "accounts.InitiateUpdateMutation", mutation)
	defer logger.Exit(ctx, "accounts.InitiateUpdateMutation", &err)

	result, err := Select(ctx, accountspg.SelectArgument{Query: mutation.Foundation.Query})
	if err != nil {
		return err
	}

	if len(result.Accounts) == 0 {
		return nil
	}

	updateAccount := mutation.Foundation.Populate(result.Accounts[0])

	mutation.Transient = accountspg.Transient{
		Current: result.Accounts[0],
		Future:  updateAccount,
	}

	return nil
}

func FinalizeUpdateMutation(ctx context.Context, mutation *accountspg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "accounts.FinalizeUpdateMutation", mutation)
	defer logger.Exit(ctx, "accounts.FinalizeUpdateMutation", &err)

	err = gormdb.Update(ctx, mutation.Transient.Future, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackUpdateMutation(ctx context.Context, mutation *accountspg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "accounts.RollbackUpdateMutation", mutation)
	defer logger.Exit(ctx, "accounts.RollbackUpdateMutation", &err)

	return FinalizeUpdateMutation(ctx, &accountspg.UpdateMutation{
		Foundation: mutation.Foundation,
		Transient: accountspg.Transient{
			Current: mutation.Transient.Current,
			Future:  mutation.Transient.Previous,
		},
	})
}
