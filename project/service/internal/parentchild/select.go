package parentchild

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
)

func Select(ctx context.Context, argument parentchildpg.SelectArgument) (res parentchildpg.SelectResponse, err error) {
	ctx = logger.Enter(ctx, "parentchild.Select", argument)
	defer logger.Exit(ctx, "parentchild.Select", &res, &err)

	var (
		usesMutation = &parentchildpg.SelectMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = ProcessSelectMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return parentchildpg.SelectResponse{
		ParentChildren: usesMutation.Transient,
	}, nil
}

func ProcessSelectMutation(ctx context.Context, mutation *parentchildpg.SelectMutation) (err error) {
	ctx = logger.Enter(ctx, "parentchild.ProcessSelectMutation", mutation)
	defer logger.Exit(ctx, "parentchild.ProcessSelectMutation", &err)

	result, err := gormdb.Select[parentchildpg.ParentChild](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = result

	return nil
}
