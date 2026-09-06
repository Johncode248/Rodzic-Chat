package absences

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	absencespg "github.com/Johncode248/Rodzic-Chat/project/common/absences"
)

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
			if err = RollbackUpdateMutation(ctx, usesMutation); err != nil {
				panic(err)
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
