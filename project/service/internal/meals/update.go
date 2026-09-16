package meals

import (
	"context"
	"errors"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	mealspg "github.com/Johncode248/Rodzic-Chat/project/common/meals"
)

// ErrInvalidDateRange mirrors CreateArgument.Validate's "date_to must not
// be before date_from" check, applied after Populate() merges the patch
// onto the existing row -- see the identical comment on
// absences.ErrInvalidDateRange for why this can't live in
// UpdateArgument.Validate() itself.
var ErrInvalidDateRange = errors.New("date_to must not be before date_from")

func Update(ctx context.Context, argument mealspg.UpdateArgument) (res mealspg.UpdateResponse, err error) {
	ctx = logger.Enter(ctx, "meals.Update", argument)
	defer logger.Exit(ctx, "meals.Update", &res, &err)

	var (
		usesMutation = &mealspg.UpdateMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = InitiateUpdateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	defer func() {
		if err != nil {
			if rollbackErr := RollbackUpdateMutation(ctx, usesMutation); rollbackErr != nil {
				panic(rollbackErr)
			}
		}
	}()

	if err = FinalizeUpdateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return mealspg.UpdateResponse{
		MealCancellation: usesMutation.Transient.Current,
	}, nil
}

func InitiateUpdateMutation(ctx context.Context, mutation *mealspg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "meals.InitiateUpdateMutation", mutation)
	defer logger.Exit(ctx, "meals.InitiateUpdateMutation", &err)

	result, err := Select(ctx, mealspg.SelectArgument{Query: mutation.Foundation.Query})
	if err != nil {
		return err
	}

	if len(result.MealCancellations) == 0 {
		return nil
	}

	updateCancellation := mutation.Foundation.Populate(result.MealCancellations[0])

	if updateCancellation.DateTo.Before(updateCancellation.DateFrom) {
		return ErrInvalidDateRange
	}

	mutation.Transient = mealspg.Transient{
		Current: result.MealCancellations[0],
		Future:  updateCancellation,
	}

	return nil
}

func FinalizeUpdateMutation(ctx context.Context, mutation *mealspg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "meals.FinalizeUpdateMutation", mutation)
	defer logger.Exit(ctx, "meals.FinalizeUpdateMutation", &err)

	err = gormdb.Update(ctx, mutation.Transient.Future, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackUpdateMutation(ctx context.Context, mutation *mealspg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "meals.RollbackUpdateMutation", mutation)
	defer logger.Exit(ctx, "meals.RollbackUpdateMutation", &err)

	if mutation.Transient.Current.Id == "" {
		return nil
	}

	return FinalizeUpdateMutation(ctx, &mealspg.UpdateMutation{
		Foundation: mutation.Foundation,
		Transient: mealspg.Transient{
			Current: mutation.Transient.Current,
			Future:  mutation.Transient.Previous,
		},
	})
}
