package chats

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
)

func SelectMessages(ctx context.Context, argument chatspg.MessageSelectArgument) (res chatspg.MessageSelectResponse, err error) {
	ctx = logger.Enter(ctx, "chats.SelectMessages", argument)
	defer logger.Exit(ctx, "chats.SelectMessages", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	result, err := gormdb.Select[chatspg.Message](ctx, argument.Query)
	if err != nil {
		return res, err
	}

	return chatspg.MessageSelectResponse{Messages: result}, nil
}
