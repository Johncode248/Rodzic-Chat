package media

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	mediapg "github.com/Johncode248/Rodzic-Chat/project/common/media"
)

func Delete(ctx context.Context, argument mediapg.DeleteArgument) (res mediapg.DeleteResponse, err error) {
	ctx = logger.Enter(ctx, "media.Delete", argument)
	defer logger.Exit(ctx, "media.Delete", &res, &err)

	var (
		usesMutation = &mediapg.DeleteMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = InitiateDeleteMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	defer func() {
		if err != nil {
			if err = RollbackDeleteMutation(ctx, usesMutation); err != nil {
				panic(err)
			}
		}
	}()

	if err = FinalizeDeleteMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	for _, value := range usesMutation.Transient {
		res.Media = append(res.Media, value.Previous)
	}

	return res, nil
}

func InitiateDeleteMutation(ctx context.Context, mutation *mediapg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "media.InitiateDeleteMutation", mutation)
	defer logger.Exit(ctx, "media.InitiateDeleteMutation", &err)

	result, err := Select(ctx, mediapg.SelectArgument{
		Query: mutation.Foundation.Query,
	})
	if err != nil {
		return err
	}

	if len(result.Media) == 0 {
		return nil
	}

	for _, item := range result.Media {
		mutation.Transient = append(mutation.Transient, mediapg.Transient{
			Current: item,
		})
	}

	return nil
}

func FinalizeDeleteMutation(ctx context.Context, mutation *mediapg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "media.FinalizeDeleteMutation", mutation)
	defer logger.Exit(ctx, "media.FinalizeDeleteMutation", &err)

	_, err = gormdb.Delete[mediapg.Media](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	for index := range mutation.Transient {
		mutation.Transient[index] = mutation.Transient[index].Update()
	}

	return nil
}

func RollbackDeleteMutation(ctx context.Context, mutation *mediapg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "media.RollbackDeleteMutation", mutation)
	defer logger.Exit(ctx, "media.RollbackDeleteMutation", &err)

	return routines.Select(
		func() error {
			functions := []func() error{}
			for _, transient := range mutation.Transient {
				if transient.Previous.Id == "" {
					continue
				}

				transient := transient
				functions = append(functions, func() error {
					return FinalizeCreateMutation(ctx, &mediapg.CreateMutation{
						Transient: mediapg.Transient{
							Future: transient.Previous,
						},
					})
				})
			}
			return routines.Select(functions...)
		},
	)
}
