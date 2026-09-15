package chats

import (
	"context"
	"errors"
	"strings"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
)

var ErrChannelNameRequired = errors.New("name is required")
var ErrChannelNameTooLong = errors.New("name must be at most 128 characters")

// RenameChannel lets an org-less parent channel's admin rename it -- "Zmień nazwę" from the
// channel's settings screen. Admin-only, same authorization shape as CreateForChannel/
// RemoveMember; a school channel's name is managed elsewhere and never reaches this function.
func RenameChannel(ctx context.Context, channelId, callerAccountId, name string) (chatspg.Channel, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return chatspg.Channel{}, ErrChannelNameRequired
	}
	if len(name) > 128 {
		return chatspg.Channel{}, ErrChannelNameTooLong
	}

	isAdmin, err := IsChannelAdmin(ctx, channelId, callerAccountId)
	if err != nil {
		return chatspg.Channel{}, err
	}
	if !isAdmin {
		return chatspg.Channel{}, ErrNotChannelAdmin
	}

	channels, err := gormdb.Select[chatspg.Channel](ctx, databasepg.Query{
		Limit:  1,
		Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: channelId}}}},
	})
	if err != nil {
		return chatspg.Channel{}, err
	}
	if len(channels) == 0 {
		return chatspg.Channel{}, errors.New("channel not found")
	}

	channel := channels[0]
	channel.Name = name
	channel.UpdateTime = time.Now()

	if err = gormdb.Update(ctx, channel, databasepg.Query{
		Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: channelId}}}},
	}); err != nil {
		return chatspg.Channel{}, err
	}

	return channel, nil
}
