package chats

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
)

func CreateChannel(ctx context.Context, argument chatspg.ChannelCreateArgument) (res chatspg.ChannelCreateResponse, err error) {
	ctx = logger.Enter(ctx, "chats.CreateChannel", argument)
	defer logger.Exit(ctx, "chats.CreateChannel", &res, &err)

	var (
		usesMutation = &chatspg.ChannelCreateMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = InitiateChannelCreateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	defer func() {
		if err != nil {
			if err = RollbackChannelCreateMutation(ctx, usesMutation); err != nil {
				panic(err)
			}
		}
	}()

	if err = FinalizeChannelCreateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return chatspg.ChannelCreateResponse{
		Channel: usesMutation.Transient.Current,
	}, nil
}

func InitiateChannelCreateMutation(ctx context.Context, mutation *chatspg.ChannelCreateMutation) (err error) {
	ctx = logger.Enter(ctx, "chats.InitiateChannelCreateMutation", mutation)
	defer logger.Exit(ctx, "chats.InitiateChannelCreateMutation", &err)

	mutation.Transient.Future = mutation.Foundation.Populate()

	return nil
}

func FinalizeChannelCreateMutation(ctx context.Context, mutation *chatspg.ChannelCreateMutation) (err error) {
	ctx = logger.Enter(ctx, "chats.FinalizeChannelCreateMutation", mutation)
	defer logger.Exit(ctx, "chats.FinalizeChannelCreateMutation", &err)

	if err = gormdb.Create(ctx, mutation.Transient.Future); err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackChannelCreateMutation(ctx context.Context, mutation *chatspg.ChannelCreateMutation) (err error) {
	ctx = logger.Enter(ctx, "chats.RollbackChannelCreateMutation", mutation)
	defer logger.Exit(ctx, "chats.RollbackChannelCreateMutation", &err)

	return routines.Select(
		func() (err error) {
			return FinalizeChannelDeleteMutation(ctx, &chatspg.ChannelDeleteMutation{
				Foundation: chatspg.ChannelDeleteArgument{
					Query: databasepg.Query{
						Groups: []databasepg.Group{
							{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: mutation.Transient.Current.Id}}},
						},
					},
				},
			})
		},
	)
}
