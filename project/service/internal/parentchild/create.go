package parentchild

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
)

func Create(ctx context.Context, argument parentchildpg.CreateArgument) (res parentchildpg.CreateResponse, err error) {
	ctx = logger.Enter(ctx, "parentchild.Create", argument)
	defer logger.Exit(ctx, "parentchild.Create", &res, &err)

	var (
		usesMutation = &parentchildpg.CreateMutation{
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
			if rollbackErr := RollbackCreateMutation(ctx, usesMutation); rollbackErr != nil {
				panic(rollbackErr)
			}
		}
	}()

	if err = FinalizeCreateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return parentchildpg.CreateResponse{
		ParentChild: usesMutation.Transient.Current,
	}, nil
}

func InitiateCreateMutation(ctx context.Context, mutation *parentchildpg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "parentchild.CreateInitiator", mutation)
	defer logger.Exit(ctx, "parentchild.CreateInitiator", &err)

	future := mutation.Foundation.Populate()

	mutation.Transient.Future = future

	return nil
}

func ValidateCreateMutation(ctx context.Context, mutation *parentchildpg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "parentchild.ValidateCreateMutation", mutation)
	defer logger.Exit(ctx, "parentchild.ValidateCreateMutation", &err)

	return nil
}

func FinalizeCreateMutation(ctx context.Context, mutation *parentchildpg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "parentchild.FinalizeCreateMutation", mutation)
	defer logger.Exit(ctx, "parentchild.FinalizeCreateMutation", &err)

	err = gormdb.Create(ctx, mutation.Transient.Future)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackCreateMutation(ctx context.Context, mutation *parentchildpg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "parentchild.RollbackCreateMutation", mutation)
	defer logger.Exit(ctx, "parentchild.RollbackCreateMutation", &err)

	return routines.Select(
		func() (err error) {
			return FinalizeDeleteMutation(ctx, &parentchildpg.DeleteMutation{
				Foundation: parentchildpg.DeleteArgument{
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
