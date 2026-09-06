package chats

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
)

func SelectChannels(ctx context.Context, argument chatspg.ChannelSelectArgument) (res chatspg.ChannelSelectResponse, err error) {
	ctx = logger.Enter(ctx, "chats.SelectChannels", argument)
	defer logger.Exit(ctx, "chats.SelectChannels", &res, &err)

	var (
		usesMutation = &chatspg.ChannelSelectMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	result, err := gormdb.Select[chatspg.Channel](ctx, usesMutation.Foundation.Query)
	if err != nil {
		return res, err
	}
	usesMutation.Transient = result

	return chatspg.ChannelSelectResponse{
		Channels: usesMutation.Transient,
	}, nil
}
