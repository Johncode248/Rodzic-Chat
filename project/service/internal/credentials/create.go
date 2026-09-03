package credentials

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	credentialspg "github.com/Johncode248/Rodzic-Chat/project/common/credentials"
)

func Create(ctx context.Context, argument credentialspg.CreateArgument) (res credentialspg.CreateResponse, err error) {
	ctx = logger.Enter(ctx, "credentials.Create", argument)
	defer logger.Exit(ctx, "credentials.Create", &res, &err)

	var (
		usesMutation = &credentialspg.CreateMutation{
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

	return credentialspg.CreateResponse{
		Credential: usesMutation.Transient.Current,
	}, nil
}

func InitiateCreateMutation(ctx context.Context, mutation *credentialspg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "credentials.InitiateCreateMutation", mutation)
	defer logger.Exit(ctx, "credentials.InitiateCreateMutation", &err)

	future := mutation.Foundation.Populate()

	mutation.Transient.Future = future

	return nil
}

func ValidateCreateMutation(ctx context.Context, mutation *credentialspg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "credentials.ValidateCreateMutation", mutation)
	defer logger.Exit(ctx, "credentials.ValidateCreateMutation", &err)

	return nil
}

func FinalizeCreateMutation(ctx context.Context, mutation *credentialspg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "credentials.FinalizeCreateMutation", mutation)
	defer logger.Exit(ctx, "credentials.FinalizeCreateMutation", &err)

	err = gormdb.Create(ctx, mutation.Transient.Future)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackCreateMutation(ctx context.Context, mutation *credentialspg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "credentials.RollbackCreateMutation", mutation)
	defer logger.Exit(ctx, "credentials.RollbackCreateMutation", &err)

	return routines.Select(
		func() (err error) {
			return ProcessDeleteMutation(ctx, &credentialspg.DeleteMutation{
				Foundation: credentialspg.DeleteArgument{
					Query: databasepg.Query{
						Groups: []databasepg.Group{
							{
								Filters: []databasepg.Filter{
									{
										Key:      "userId",
										Operator: "=",
										Value:    mutation.Transient.Current.UserId,
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
