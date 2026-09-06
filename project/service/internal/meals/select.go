package meals

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	mealspg "github.com/Johncode248/Rodzic-Chat/project/common/meals"
)

func Select(ctx context.Context, argument mealspg.SelectArgument) (res mealspg.SelectResponse, err error) {
	ctx = logger.Enter(ctx, "meals.Select", argument)
	defer logger.Exit(ctx, "meals.Select", &res, &err)

	var (
		usesMutation = &mealspg.SelectMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = ProcessSelectMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return mealspg.SelectResponse{
		MealCancellations: usesMutation.Transient,
	}, nil
}

func ProcessSelectMutation(ctx context.Context, mutation *mealspg.SelectMutation) (err error) {
	ctx = logger.Enter(ctx, "meals.ProcessSelectMutation", mutation)
	defer logger.Exit(ctx, "meals.ProcessSelectMutation", &err)

	result, err := gormdb.Select[mealspg.MealCancellation](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = result

	return nil
}
