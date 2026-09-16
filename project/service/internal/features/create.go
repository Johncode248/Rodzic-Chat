package features

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	featurespg "github.com/Johncode248/Rodzic-Chat/project/common/features"
)

func Create(ctx context.Context, argument featurespg.CreateArgument) (res featurespg.CreateResponse, err error) {
	ctx = logger.Enter(ctx, "features.Create", argument)
	defer logger.Exit(ctx, "features.Create", &res, &err)

	var (
		usesMutation = &featurespg.CreateMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = InitiateCreateMutation(ctx, usesMutation); err != nil {
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

	return featurespg.CreateResponse{
		OrganizationFeature: usesMutation.Transient.Current,
	}, nil
}

func InitiateCreateMutation(ctx context.Context, mutation *featurespg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "features.InitiateCreateMutation", mutation)
	defer logger.Exit(ctx, "features.InitiateCreateMutation", &err)

	future := mutation.Foundation.Populate()

	mutation.Transient.Future = future

	return nil
}

func FinalizeCreateMutation(ctx context.Context, mutation *featurespg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "features.FinalizeCreateMutation", mutation)
	defer logger.Exit(ctx, "features.FinalizeCreateMutation", &err)

	err = gormdb.Create(ctx, mutation.Transient.Future)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackCreateMutation(ctx context.Context, mutation *featurespg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "features.RollbackCreateMutation", mutation)
	defer logger.Exit(ctx, "features.RollbackCreateMutation", &err)

	return routines.Select(
		func() (err error) {
			return FinalizeDeleteMutation(ctx, &featurespg.DeleteMutation{
				Foundation: featurespg.DeleteArgument{
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
