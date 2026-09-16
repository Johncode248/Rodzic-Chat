package absences

import (
	"context"
	"errors"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	absencespg "github.com/Johncode248/Rodzic-Chat/project/common/absences"
)

// ErrInvalidDateRange mirrors CreateArgument.Validate's "date_to must not
// be before date_from" check -- Update's DateFrom/DateTo are independently
// optional (a partial patch), so the same check can't live in
// UpdateArgument.Validate() itself (it only sees whichever of the two
// fields were actually sent, not the existing row's other date); it has to
// run after Populate() below merges the patch onto the existing row.
var ErrInvalidDateRange = errors.New("date_to must not be before date_from")

func Update(ctx context.Context, argument absencespg.UpdateArgument) (res absencespg.UpdateResponse, err error) {
	ctx = logger.Enter(ctx, "absences.Update", argument)
	defer logger.Exit(ctx, "absences.Update", &res, &err)

	var (
		usesMutation = &absencespg.UpdateMutation{
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

	return absencespg.UpdateResponse{
		Absence: usesMutation.Transient.Current,
	}, nil
}

func InitiateUpdateMutation(ctx context.Context, mutation *absencespg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "absences.InitiateUpdateMutation", mutation)
	defer logger.Exit(ctx, "absences.InitiateUpdateMutation", &err)

	result, err := Select(ctx, absencespg.SelectArgument{Query: mutation.Foundation.Query})
	if err != nil {
		return err
	}

	if len(result.Absences) == 0 {
		return nil
	}

	updateAbsence := mutation.Foundation.Populate(result.Absences[0])

	if updateAbsence.DateTo.Before(updateAbsence.DateFrom) {
		return ErrInvalidDateRange
	}

	mutation.Transient = absencespg.Transient{
		Current: result.Absences[0],
		Future:  updateAbsence,
	}

	return nil
}

func FinalizeUpdateMutation(ctx context.Context, mutation *absencespg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "absences.FinalizeUpdateMutation", mutation)
	defer logger.Exit(ctx, "absences.FinalizeUpdateMutation", &err)

	err = gormdb.Update(ctx, mutation.Transient.Future, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackUpdateMutation(ctx context.Context, mutation *absencespg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "absences.RollbackUpdateMutation", mutation)
	defer logger.Exit(ctx, "absences.RollbackUpdateMutation", &err)

	if mutation.Transient.Current.Id == "" {
		return nil
	}

	return FinalizeUpdateMutation(ctx, &absencespg.UpdateMutation{
		Foundation: mutation.Foundation,
		Transient: absencespg.Transient{
			Current: mutation.Transient.Current,
			Future:  mutation.Transient.Previous,
		},
	})
}
