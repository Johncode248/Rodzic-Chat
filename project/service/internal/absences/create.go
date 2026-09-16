package absences

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	absencespg "github.com/Johncode248/Rodzic-Chat/project/common/absences"
)

func Create(ctx context.Context, argument absencespg.CreateArgument) (res absencespg.CreateResponse, err error) {
	ctx = logger.Enter(ctx, "absences.Create", argument)
	defer logger.Exit(ctx, "absences.Create", &res, &err)

	var (
		usesMutation = &absencespg.CreateMutation{
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

	return absencespg.CreateResponse{
		Absence: usesMutation.Transient.Current,
	}, nil
}

func InitiateCreateMutation(ctx context.Context, mutation *absencespg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "absences.InitiateCreateMutation", mutation)
	defer logger.Exit(ctx, "absences.InitiateCreateMutation", &err)

	future := mutation.Foundation.Populate()

	mutation.Transient.Future = future

	return nil
}

func FinalizeCreateMutation(ctx context.Context, mutation *absencespg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "absences.FinalizeCreateMutation", mutation)
	defer logger.Exit(ctx, "absences.FinalizeCreateMutation", &err)

	err = gormdb.Create(ctx, mutation.Transient.Future)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackCreateMutation(ctx context.Context, mutation *absencespg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "absences.RollbackCreateMutation", mutation)
	defer logger.Exit(ctx, "absences.RollbackCreateMutation", &err)

	return routines.Select(
		func() (err error) {
			return FinalizeDeleteMutation(ctx, &absencespg.DeleteMutation{
				Foundation: absencespg.DeleteArgument{
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
