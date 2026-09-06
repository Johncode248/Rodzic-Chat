package groups

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	groupspg "github.com/Johncode248/Rodzic-Chat/project/common/groups"
)

func Select(ctx context.Context, argument groupspg.SelectArgument) (res groupspg.SelectResponse, err error) {
	ctx = logger.Enter(ctx, "groups.Select", argument)
	defer logger.Exit(ctx, "groups.Select", &res, &err)

	var (
		usesMutation = &groupspg.SelectMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = ProcessSelectMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return groupspg.SelectResponse{
		Groups: usesMutation.Transient,
	}, nil
}

func ProcessSelectMutation(ctx context.Context, mutation *groupspg.SelectMutation) (err error) {
	ctx = logger.Enter(ctx, "groups.ProcessSelectMutation", mutation)
	defer logger.Exit(ctx, "groups.ProcessSelectMutation", &err)

	result, err := gormdb.Select[groupspg.Group](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = result

	return nil
}
