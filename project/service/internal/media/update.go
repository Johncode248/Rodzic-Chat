package media

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	mediapg "github.com/Johncode248/Rodzic-Chat/project/common/media"
)

func Update(ctx context.Context, argument mediapg.UpdateArgument) (res mediapg.UpdateResponse, err error) {
	ctx = logger.Enter(ctx, "media.Update", argument)
	defer logger.Exit(ctx, "media.Update", &res, &err)

	var (
		usesMutation = &mediapg.UpdateMutation{
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

	return mediapg.UpdateResponse{
		Media: usesMutation.Transient.Current,
	}, nil
}

func InitiateUpdateMutation(ctx context.Context, mutation *mediapg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "media.InitiateUpdateMutation", mutation)
	defer logger.Exit(ctx, "media.InitiateUpdateMutation", &err)

	result, err := Select(ctx, mediapg.SelectArgument{Query: mutation.Foundation.Query})
	if err != nil {
		return err
	}

	if len(result.Media) == 0 {
		return nil
	}

	updateMedia := mutation.Foundation.Populate(result.Media[0])

	mutation.Transient = mediapg.Transient{
		Current: result.Media[0],
		Future:  updateMedia,
	}

	return nil
}

func FinalizeUpdateMutation(ctx context.Context, mutation *mediapg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "media.FinalizeUpdateMutation", mutation)
	defer logger.Exit(ctx, "media.FinalizeUpdateMutation", &err)

	err = gormdb.Update(ctx, mutation.Transient.Future, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackUpdateMutation(ctx context.Context, mutation *mediapg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "media.RollbackUpdateMutation", mutation)
	defer logger.Exit(ctx, "media.RollbackUpdateMutation", &err)

	if mutation.Transient.Current.Id == "" {
		return nil
	}

	return FinalizeUpdateMutation(ctx, &mediapg.UpdateMutation{
		Foundation: mutation.Foundation,
		Transient: mediapg.Transient{
			Current: mutation.Transient.Current,
			Future:  mutation.Transient.Previous,
		},
	})
}
