package chats

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
)

func DeleteVote(ctx context.Context, argument chatspg.PollVoteDeleteArgument) (res chatspg.PollVoteDeleteResponse, err error) {
	ctx = logger.Enter(ctx, "chats.DeleteVote", argument)
	defer logger.Exit(ctx, "chats.DeleteVote", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	result, err := gormdb.Delete[chatspg.PollVote](ctx, argument.Query)
	if err != nil {
		return res, err
	}

	return chatspg.PollVoteDeleteResponse{PollVotes: result}, nil
}
