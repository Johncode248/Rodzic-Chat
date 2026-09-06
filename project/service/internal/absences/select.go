package absences

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	absencespg "github.com/Johncode248/Rodzic-Chat/project/common/absences"
)

func Select(ctx context.Context, argument absencespg.SelectArgument) (res absencespg.SelectResponse, err error) {
	ctx = logger.Enter(ctx, "absences.Select", argument)
	defer logger.Exit(ctx, "absences.Select", &res, &err)

	var (
		usesMutation = &absencespg.SelectMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = ProcessSelectMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return absencespg.SelectResponse{
		Absences: usesMutation.Transient,
	}, nil
}

func ProcessSelectMutation(ctx context.Context, mutation *absencespg.SelectMutation) (err error) {
	ctx = logger.Enter(ctx, "absences.ProcessSelectMutation", mutation)
	defer logger.Exit(ctx, "absences.ProcessSelectMutation", &err)

	result, err := gormdb.Select[absencespg.Absence](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = result

	return nil
}
