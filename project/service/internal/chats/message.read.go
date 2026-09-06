package chats

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
)

// MarkRead is an upsert on (message_id, account_id): re-marking a message
// read is idempotent and just bumps ReadAt on the existing row.
func MarkRead(ctx context.Context, argument chatspg.MessageReadCreateArgument) (res chatspg.MessageReadCreateResponse, err error) {
	ctx = logger.Enter(ctx, "chats.MarkRead", argument)
	defer logger.Exit(ctx, "chats.MarkRead", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	query := databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{
				{Key: "message_id", Operator: "=", Value: argument.MessageId},
				{Key: "account_id", Operator: "=", Value: argument.AccountId},
			}},
		},
	}

	existing, err := gormdb.Select[chatspg.MessageRead](ctx, query)
	if err != nil {
		return res, err
	}

	read := argument.Populate()

	if len(existing) > 0 {
		read.Id = existing[0].Id
		if err = gormdb.Update(ctx, read, query); err != nil {
			return res, err
		}
		return chatspg.MessageReadCreateResponse{MessageRead: read}, nil
	}

	if err = gormdb.Create(ctx, read); err != nil {
		return res, err
	}

	return chatspg.MessageReadCreateResponse{MessageRead: read}, nil
}

func SelectMessageReads(ctx context.Context, argument chatspg.MessageReadSelectArgument) (res chatspg.MessageReadSelectResponse, err error) {
	ctx = logger.Enter(ctx, "chats.SelectMessageReads", argument)
	defer logger.Exit(ctx, "chats.SelectMessageReads", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	result, err := gormdb.Select[chatspg.MessageRead](ctx, argument.Query)
	if err != nil {
		return res, err
	}

	return chatspg.MessageReadSelectResponse{MessageReads: result}, nil
}
