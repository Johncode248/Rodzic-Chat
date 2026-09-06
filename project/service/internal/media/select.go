package media

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	mediapg "github.com/Johncode248/Rodzic-Chat/project/common/media"
)

func Select(ctx context.Context, argument mediapg.SelectArgument) (res mediapg.SelectResponse, err error) {
	ctx = logger.Enter(ctx, "media.Select", argument)
	defer logger.Exit(ctx, "media.Select", &res, &err)

	var (
		usesMutation = &mediapg.SelectMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = ProcessSelectMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return mediapg.SelectResponse{
		Media: usesMutation.Transient,
	}, nil
}

func ProcessSelectMutation(ctx context.Context, mutation *mediapg.SelectMutation) (err error) {
	ctx = logger.Enter(ctx, "media.ProcessSelectMutation", mutation)
	defer logger.Exit(ctx, "media.ProcessSelectMutation", &err)

	result, err := gormdb.Select[mediapg.Media](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = result

	return nil
}
