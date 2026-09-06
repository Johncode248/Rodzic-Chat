package teacherinvites

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	teacherinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/teacherinvites"
)

func Select(ctx context.Context, argument teacherinvitespg.SelectArgument) (res teacherinvitespg.SelectResponse, err error) {
	ctx = logger.Enter(ctx, "teacherinvites.Select", argument)
	defer logger.Exit(ctx, "teacherinvites.Select", &res, &err)

	var (
		usesMutation = &teacherinvitespg.SelectMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = ProcessSelectMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return teacherinvitespg.SelectResponse{
		TeacherInvites: usesMutation.Transient,
	}, nil
}

func ProcessSelectMutation(ctx context.Context, mutation *teacherinvitespg.SelectMutation) (err error) {
	ctx = logger.Enter(ctx, "teacherinvites.ProcessSelectMutation", mutation)
	defer logger.Exit(ctx, "teacherinvites.ProcessSelectMutation", &err)

	result, err := gormdb.Select[teacherinvitespg.TeacherInvite](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = result

	return nil
}
