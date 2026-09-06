package accounts

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
)

func Create(ctx context.Context, argument accountspg.CreateArgument) (res accountspg.CreateResponse, err error) {
	ctx = logger.Enter(ctx, "accounts.Create", argument)
	defer logger.Exit(ctx, "accounts.Create", &res, &err)

	var (
		usesMutation = &accountspg.CreateMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = InitiateCreateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	if err = ValidateCreateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	defer func() {
		if err != nil {
			if err = RollbackCreateMutation(ctx, usesMutation); err != nil {
				panic(err)
			}
		}
	}()

	if err = FinalizeCreateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return accountspg.CreateResponse{
		Account: usesMutation.Transient.Current,
	}, nil
}

func InitiateCreateMutation(ctx context.Context, mutation *accountspg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "accounts.CreateInitiator", mutation)
	defer logger.Exit(ctx, "accounts.CreateInitiator", &err)

	future := mutation.Foundation.Populate()

	mutation.Transient.Future = future

	return nil
}

func ValidateCreateMutation(ctx context.Context, mutation *accountspg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "accounts.ValidateCreateMutation", mutation)
	defer logger.Exit(ctx, "accounts.ValidateCreateMutation", &err)

	return nil
}

func FinalizeCreateMutation(ctx context.Context, mutation *accountspg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "accounts.FinalizeCreateMutation", mutation)
	defer logger.Exit(ctx, "accounts.FinalizeCreateMutation", &err)

	err = gormdb.Create(ctx, mutation.Transient.Future)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackCreateMutation(ctx context.Context, mutation *accountspg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "accounts.RollbackCreateMutation", mutation)
	defer logger.Exit(ctx, "accounts.RollbackCreateMutation", &err)

	return routines.Select(
		func() (err error) {
			return FinalizeDeleteMutation(ctx, &accountspg.DeleteMutation{
				Foundation: accountspg.DeleteArgument{
					Query: databasepg.Query{
						Groups: []databasepg.Group{
							{
								Filters: []databasepg.Filter{
									{
										Key:      "id",
										Operator: "=",
										Value:    mutation.Transient.Current.Id,
									},
								},
							},
						},
					},
				},
			})
		},
	)
}
