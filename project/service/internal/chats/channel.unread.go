package chats

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
)

// AttachUnreadFlags sets Channel.Unread in place for each channel: true when the channel's most
// recent message has no MessageRead row for accountId. Read tracking is per-message only (no
// channel-level read cursor exists), so "unread" is derived rather than stored -- a channel with
// no messages at all is never unread. Called by both the org-scoped and org-less channel-list
// handlers, which are the only two call sites that know who's actually asking.
func AttachUnreadFlags(ctx context.Context, channels []chatspg.Channel, accountId string) ([]chatspg.Channel, error) {
	for i, channel := range channels {
		messages, err := gormdb.Select[chatspg.Message](ctx, databasepg.Query{
			Limit:   1,
			Groups:  []databasepg.Group{{Filters: []databasepg.Filter{{Key: "channel_id", Operator: "=", Value: channel.Id}}}},
			OrderBy: databasepg.Order{Key: "create_time", Direction: databasepg.DirectionDescending},
		})
		if err != nil {
			return nil, err
		}
		if len(messages) == 0 {
			continue
		}

		reads, err := gormdb.Select[chatspg.MessageRead](ctx, databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{
				{Key: "message_id", Operator: "=", Value: messages[0].Id},
				{Key: "account_id", Operator: "=", Value: accountId},
			}}},
		})
		if err != nil {
			return nil, err
		}
		channels[i].Unread = len(reads) == 0
	}
	return channels, nil
}
