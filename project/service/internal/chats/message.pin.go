package chats

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
)

var ErrMessageNotFound = errors.New("message not found")

// LoadMessage returns a single message by id, or ErrMessageNotFound.
func LoadMessage(ctx context.Context, messageId string) (chatspg.Message, error) {
	result, err := SelectMessages(ctx, chatspg.MessageSelectArgument{
		Query: databasepg.Query{
			Limit:  1,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: messageId}}}},
		},
	})
	if err != nil {
		return chatspg.Message{}, err
	}
	if len(result.Messages) == 0 {
		return chatspg.Message{}, ErrMessageNotFound
	}
	return result.Messages[0], nil
}

// SetMessagePinned pins or unpins messageId on behalf of accountId. Authorization (is the caller
// allowed to touch this channel at all) is the transport handler's job; this only persists the
// change and returns the updated row.
func SetMessagePinned(ctx context.Context, messageId, accountId string, pinned bool) (res chatspg.Message, err error) {
	ctx = logger.Enter(ctx, "chats.SetMessagePinned", messageId, pinned)
	defer logger.Exit(ctx, "chats.SetMessagePinned", &res, &err)

	if _, err = LoadMessage(ctx, messageId); err != nil {
		return res, err
	}

	updates := map[string]interface{}{"pinned_at": nil, "pinned_by": ""}
	if pinned {
		updates["pinned_at"] = time.Now()
		updates["pinned_by"] = accountId
	}
	if err = gormdb.GormDB.WithContext(ctx).Model(&chatspg.Message{}).Where("id = ?", messageId).Updates(updates).Error; err != nil {
		return res, err
	}

	return LoadMessage(ctx, messageId)
}
