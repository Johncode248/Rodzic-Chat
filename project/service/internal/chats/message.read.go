package chats

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"gorm.io/gorm/clause"
)

// MarkRead is an upsert on (message_id, account_id): re-marking a message
// read is idempotent and just bumps ReadAt on the existing row.
//
// This has to be a single atomic upsert, not select-then-branch: two
// concurrent calls for the same (message_id, account_id) -- e.g. a double
// click, or the client re-sending on a flaky connection -- would otherwise
// both see "no existing row" from the SELECT, both attempt an INSERT, and
// the second would hit MessageRead's unique (message_id, account_id) index
// and fail with a raw DB constraint error (surfaced to the client as an
// HTTP 500) instead of the idempotent "already marked read" it should be.
func MarkRead(ctx context.Context, argument chatspg.MessageReadCreateArgument) (res chatspg.MessageReadCreateResponse, err error) {
	ctx = logger.Enter(ctx, "chats.MarkRead", argument)
	defer logger.Exit(ctx, "chats.MarkRead", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	read := argument.Populate()

	if err = gormdb.GormDB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "message_id"}, {Name: "account_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"read_at"}),
	}).Create(&read).Error; err != nil {
		return res, err
	}

	// Re-read the canonical row rather than trust `read` in memory: on a
	// conflict, the row actually written keeps its original Id (only
	// read_at is updated), which the in-memory struct above -- built with a
	// freshly generated Id from Populate() -- does not necessarily reflect.
	existing, err := gormdb.Select[chatspg.MessageRead](ctx, databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{
				{Key: "message_id", Operator: "=", Value: argument.MessageId},
				{Key: "account_id", Operator: "=", Value: argument.AccountId},
			}},
		},
	})
	if err != nil {
		return res, err
	}
	if len(existing) > 0 {
		read = existing[0]
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
