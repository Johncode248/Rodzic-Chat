package children

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
)

func Create(ctx context.Context, argument childrenpg.CreateArgument) (res childrenpg.CreateResponse, err error) {
	ctx = logger.Enter(ctx, "children.Create", argument)
	defer logger.Exit(ctx, "children.Create", &res, &err)

	var (
		usesMutation = &childrenpg.CreateMutation{
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

	return childrenpg.CreateResponse{
		Child: usesMutation.Transient.Current,
	}, nil
}

func InitiateCreateMutation(ctx context.Context, mutation *childrenpg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "children.CreateInitiator", mutation)
	defer logger.Exit(ctx, "children.CreateInitiator", &err)

	future := mutation.Foundation.Populate()

	mutation.Transient.Future = future

	return nil
}

func ValidateCreateMutation(ctx context.Context, mutation *childrenpg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "children.ValidateCreateMutation", mutation)
	defer logger.Exit(ctx, "children.ValidateCreateMutation", &err)

	return nil
}

func FinalizeCreateMutation(ctx context.Context, mutation *childrenpg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "children.FinalizeCreateMutation", mutation)
	defer logger.Exit(ctx, "children.FinalizeCreateMutation", &err)

	err = gormdb.Create(ctx, mutation.Transient.Future)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackCreateMutation(ctx context.Context, mutation *childrenpg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "children.RollbackCreateMutation", mutation)
	defer logger.Exit(ctx, "children.RollbackCreateMutation", &err)

	return routines.Select(
		func() (err error) {
			return FinalizeDeleteMutation(ctx, &childrenpg.DeleteMutation{
				Foundation: childrenpg.DeleteArgument{
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
