package organizations

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
)

func Create(ctx context.Context, argument organizationspg.CreateArgument) (res organizationspg.CreateResponse, err error) {
	ctx = logger.Enter(ctx, "organizations.Create", argument)
	defer logger.Exit(ctx, "organizations.Create", &res, &err)

	var (
		usesMutation = &organizationspg.CreateMutation{
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

	return organizationspg.CreateResponse{
		Organization: usesMutation.Transient.Current,
	}, nil
}

func InitiateCreateMutation(ctx context.Context, mutation *organizationspg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "organizations.CreateInitiator", mutation)
	defer logger.Exit(ctx, "organizations.CreateInitiator", &err)

	future := mutation.Foundation.Populate()

	mutation.Transient.Future = future

	return nil
}

func ValidateCreateMutation(ctx context.Context, mutation *organizationspg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "organizations.ValidateCreateMutation", mutation)
	defer logger.Exit(ctx, "organizations.ValidateCreateMutation", &err)

	return nil
}

func FinalizeCreateMutation(ctx context.Context, mutation *organizationspg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "organizations.FinalizeCreateMutation", mutation)
	defer logger.Exit(ctx, "organizations.FinalizeCreateMutation", &err)

	err = gormdb.Create(ctx, mutation.Transient.Future)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackCreateMutation(ctx context.Context, mutation *organizationspg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "organizations.RollbackCreateMutation", mutation)
	defer logger.Exit(ctx, "organizations.RollbackCreateMutation", &err)

	return routines.Select(
		func() (err error) {
			return FinalizeDeleteMutation(ctx, &organizationspg.DeleteMutation{
				Foundation: organizationspg.DeleteArgument{
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
