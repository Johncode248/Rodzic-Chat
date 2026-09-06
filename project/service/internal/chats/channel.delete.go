package chats

import (
	"context"
	"errors"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
)

var ErrParentsChannelNotDeletable = errors.New("the default parents channel cannot be deleted")

func DeleteChannel(ctx context.Context, argument chatspg.ChannelDeleteArgument) (res chatspg.ChannelDeleteResponse, err error) {
	ctx = logger.Enter(ctx, "chats.DeleteChannel", argument)
	defer logger.Exit(ctx, "chats.DeleteChannel", &res, &err)

	var (
		usesMutation = &chatspg.ChannelDeleteMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = InitiateChannelDeleteMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	defer func() {
		if err != nil {
			if err = RollbackChannelDeleteMutation(ctx, usesMutation); err != nil {
				panic(err)
			}
		}
	}()

	if err = FinalizeChannelDeleteMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	for _, value := range usesMutation.Transient {
		res.Channels = append(res.Channels, value.Previous)
	}

	return res, nil
}

func InitiateChannelDeleteMutation(ctx context.Context, mutation *chatspg.ChannelDeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "chats.InitiateChannelDeleteMutation", mutation)
	defer logger.Exit(ctx, "chats.InitiateChannelDeleteMutation", &err)

	result, err := SelectChannels(ctx, chatspg.ChannelSelectArgument{Query: mutation.Foundation.Query})
	if err != nil {
		return err
	}

	for _, channel := range result.Channels {
		if channel.Type == chatspg.ChannelTypeParents {
			return ErrParentsChannelNotDeletable
		}
		mutation.Transient = append(mutation.Transient, chatspg.ChannelTransient{Current: channel})
	}

	return nil
}

func FinalizeChannelDeleteMutation(ctx context.Context, mutation *chatspg.ChannelDeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "chats.FinalizeChannelDeleteMutation", mutation)
	defer logger.Exit(ctx, "chats.FinalizeChannelDeleteMutation", &err)

	if _, err = gormdb.Delete[chatspg.Channel](ctx, mutation.Foundation.Query); err != nil {
		return err
	}

	for index := range mutation.Transient {
		mutation.Transient[index] = mutation.Transient[index].Update()
	}

	return nil
}

func RollbackChannelDeleteMutation(ctx context.Context, mutation *chatspg.ChannelDeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "chats.RollbackChannelDeleteMutation", mutation)
	defer logger.Exit(ctx, "chats.RollbackChannelDeleteMutation", &err)

	return routines.Select(
		func() error {
			functions := []func() error{}
			for _, transient := range mutation.Transient {
				if transient.Previous.Id == "" {
					continue
				}
				transient := transient
				functions = append(functions, func() error {
					return FinalizeChannelCreateMutation(ctx, &chatspg.ChannelCreateMutation{
						Transient: chatspg.ChannelTransient{Future: transient.Previous},
					})
				})
			}
			return routines.Select(functions...)
		},
	)
}
