package meals

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	mealspg "github.com/Johncode248/Rodzic-Chat/project/common/meals"
)

func Create(ctx context.Context, argument mealspg.CreateArgument) (res mealspg.CreateResponse, err error) {
	ctx = logger.Enter(ctx, "meals.Create", argument)
	defer logger.Exit(ctx, "meals.Create", &res, &err)

	var (
		usesMutation = &mealspg.CreateMutation{
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
			if err = RollbackCreateMutation(ctx, usesMutation); err != nil {
				panic(err)
			}
		}
	}()

	if err = FinalizeCreateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return mealspg.CreateResponse{
		MealCancellation: usesMutation.Transient.Current,
	}, nil
}

func InitiateCreateMutation(ctx context.Context, mutation *mealspg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "meals.InitiateCreateMutation", mutation)
	defer logger.Exit(ctx, "meals.InitiateCreateMutation", &err)

	future := mutation.Foundation.Populate()

	mutation.Transient.Future = future

	return nil
}

func FinalizeCreateMutation(ctx context.Context, mutation *mealspg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "meals.FinalizeCreateMutation", mutation)
	defer logger.Exit(ctx, "meals.FinalizeCreateMutation", &err)

	err = gormdb.Create(ctx, mutation.Transient.Future)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackCreateMutation(ctx context.Context, mutation *mealspg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "meals.RollbackCreateMutation", mutation)
	defer logger.Exit(ctx, "meals.RollbackCreateMutation", &err)

	return routines.Select(
		func() (err error) {
			return FinalizeDeleteMutation(ctx, &mealspg.DeleteMutation{
				Foundation: mealspg.DeleteArgument{
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
