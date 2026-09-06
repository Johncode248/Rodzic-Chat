package meals

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	mealspg "github.com/Johncode248/Rodzic-Chat/project/common/meals"
)

func Delete(ctx context.Context, argument mealspg.DeleteArgument) (res mealspg.DeleteResponse, err error) {
	ctx = logger.Enter(ctx, "meals.Delete", argument)
	defer logger.Exit(ctx, "meals.Delete", &res, &err)

	var (
		usesMutation = &mealspg.DeleteMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = InitiateDeleteMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	defer func() {
		if err != nil {
			if err = RollbackDeleteMutation(ctx, usesMutation); err != nil {
				panic(err)
			}
		}
	}()

	if err = FinalizeDeleteMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	for _, value := range usesMutation.Transient {
		res.MealCancellations = append(res.MealCancellations, value.Previous)
	}

	return res, nil
}

func InitiateDeleteMutation(ctx context.Context, mutation *mealspg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "meals.InitiateDeleteMutation", mutation)
	defer logger.Exit(ctx, "meals.InitiateDeleteMutation", &err)

	result, err := Select(ctx, mealspg.SelectArgument{
		Query: mutation.Foundation.Query,
	})
	if err != nil {
		return err
	}

	if len(result.MealCancellations) == 0 {
		return nil
	}

	for _, cancellation := range result.MealCancellations {
		mutation.Transient = append(mutation.Transient, mealspg.Transient{
			Current: cancellation,
		})
	}

	return nil
}

func FinalizeDeleteMutation(ctx context.Context, mutation *mealspg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "meals.FinalizeDeleteMutation", mutation)
	defer logger.Exit(ctx, "meals.FinalizeDeleteMutation", &err)

	_, err = gormdb.Delete[mealspg.MealCancellation](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	for index := range mutation.Transient {
		mutation.Transient[index] = mutation.Transient[index].Update()
	}

	return nil
}

func RollbackDeleteMutation(ctx context.Context, mutation *mealspg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "meals.RollbackDeleteMutation", mutation)
	defer logger.Exit(ctx, "meals.RollbackDeleteMutation", &err)

	return routines.Select(
		func() error {
			functions := []func() error{}
			for _, transient := range mutation.Transient {
				if transient.Previous.Id == "" {
					continue
				}

				transient := transient
				functions = append(functions, func() error {
					return FinalizeCreateMutation(ctx, &mealspg.CreateMutation{
						Transient: mealspg.Transient{
							Future: transient.Previous,
						},
					})
				})
			}
			return routines.Select(functions...)
		},
	)
}
