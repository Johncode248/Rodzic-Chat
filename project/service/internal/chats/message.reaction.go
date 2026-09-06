package chats

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
)

// UpsertReaction replaces the caller's existing reaction on this message
// (if any) with the new emoji -- one reaction per person per message.
func UpsertReaction(ctx context.Context, argument chatspg.MessageReactionCreateArgument) (res chatspg.MessageReactionCreateResponse, err error) {
	ctx = logger.Enter(ctx, "chats.UpsertReaction", argument)
	defer logger.Exit(ctx, "chats.UpsertReaction", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if _, err = gormdb.Delete[chatspg.MessageReaction](ctx, databasepg.Query{
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{
				{Key: "message_id", Operator: "=", Value: argument.MessageId},
				{Key: "account_id", Operator: "=", Value: argument.AccountId},
			}},
		},
	}); err != nil {
		return res, err
	}

	reaction := argument.Populate()
	if err = gormdb.Create(ctx, reaction); err != nil {
		return res, err
	}

	return chatspg.MessageReactionCreateResponse{MessageReaction: reaction}, nil
}

// SelectReactions lists every reaction on a message -- unlike read receipts,
// reactions aren't privacy-sensitive (they're the equivalent of a public
// "like"), so this is open to any org member, no staff-only filtering.
func SelectReactions(ctx context.Context, argument chatspg.MessageReactionSelectArgument) (res chatspg.MessageReactionSelectResponse, err error) {
	ctx = logger.Enter(ctx, "chats.SelectReactions", argument)
	defer logger.Exit(ctx, "chats.SelectReactions", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	result, err := gormdb.Select[chatspg.MessageReaction](ctx, argument.Query)
	if err != nil {
		return res, err
	}

	return chatspg.MessageReactionSelectResponse{MessageReactions: result}, nil
}

func DeleteReaction(ctx context.Context, argument chatspg.MessageReactionDeleteArgument) (res chatspg.MessageReactionDeleteResponse, err error) {
	ctx = logger.Enter(ctx, "chats.DeleteReaction", argument)
	defer logger.Exit(ctx, "chats.DeleteReaction", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	result, err := gormdb.Delete[chatspg.MessageReaction](ctx, argument.Query)
	if err != nil {
		return res, err
	}

	return chatspg.MessageReactionDeleteResponse{MessageReactions: result}, nil
}
