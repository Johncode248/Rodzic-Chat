package media

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	mediapg "github.com/Johncode248/Rodzic-Chat/project/common/media"
)

func Create(ctx context.Context, argument mediapg.CreateArgument) (res mediapg.CreateResponse, err error) {
	ctx = logger.Enter(ctx, "media.Create", argument)
	defer logger.Exit(ctx, "media.Create", &res, &err)

	var (
		usesMutation = &mediapg.CreateMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = InitiateCreateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	defer func() {
		if err != nil {
			if err = RollbackCreateMutation(ctx, usesMutation); err != nil {
				panic(err)
			}
		}
	}()

	if err = FinalizeCreateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return mediapg.CreateResponse{
		Media: usesMutation.Transient.Current,
	}, nil
}

func InitiateCreateMutation(ctx context.Context, mutation *mediapg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "media.InitiateCreateMutation", mutation)
	defer logger.Exit(ctx, "media.InitiateCreateMutation", &err)

	future := mutation.Foundation.Populate()

	mutation.Transient.Future = future

	return nil
}

func FinalizeCreateMutation(ctx context.Context, mutation *mediapg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "media.FinalizeCreateMutation", mutation)
	defer logger.Exit(ctx, "media.FinalizeCreateMutation", &err)

	err = gormdb.Create(ctx, mutation.Transient.Future)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackCreateMutation(ctx context.Context, mutation *mediapg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "media.RollbackCreateMutation", mutation)
	defer logger.Exit(ctx, "media.RollbackCreateMutation", &err)

	return routines.Select(
		func() (err error) {
			return FinalizeDeleteMutation(ctx, &mediapg.DeleteMutation{
				Foundation: mediapg.DeleteArgument{
					Query: databasepg.Query{
						Groups: []databasepg.Group{
							{
								Filters: []databasepg.Filter{
									{
										Key:      "id",
										Operator: "=",
										Value:    mutation.Transient.Current.Id,
									},
								},
							},
						},
					},
				},
			})
		},
	)
}
