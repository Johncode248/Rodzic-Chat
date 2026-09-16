package chats

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
)

// latestMessagePerChannelRow is the scan target for AttachUnreadFlags'
// "one most recent message id per channel" query.
type latestMessagePerChannelRow struct {
	ChannelId string
	Id        string
}

// AttachUnreadFlags sets Channel.Unread in place for each channel: true when the channel's most
// recent message has no MessageRead row for accountId. Read tracking is per-message only (no
// channel-level read cursor exists), so "unread" is derived rather than stored -- a channel with
// no messages at all is never unread. Called by both the org-scoped and org-less channel-list
// handlers, which are the only two call sites that know who's actually asking.
//
// This runs in a fixed 2 queries no matter how many channels are passed in
// (a Postgres DISTINCT ON to get each channel's latest message id, then one
// IN-list lookup of which of those ids accountId has read) -- the previous
// version issued 2 queries PER channel (N+1), which meant a director's
// channel list page made 2x as many round trips as it had channels.
func AttachUnreadFlags(ctx context.Context, channels []chatspg.Channel, accountId string) ([]chatspg.Channel, error) {
	if len(channels) == 0 {
		return channels, nil
	}

	channelIds := make([]string, len(channels))
	for i, channel := range channels {
		channelIds[i] = channel.Id
	}

	var latestRows []latestMessagePerChannelRow
	if err := gormdb.GormDB.WithContext(ctx).Model(&chatspg.Message{}).
		Select("DISTINCT ON (channel_id) channel_id, id").
		Where("channel_id IN ?", channelIds).
		Order("channel_id, create_time DESC").
		Scan(&latestRows).Error; err != nil {
		return nil, err
	}

	latestMessageIdByChannel := make(map[string]string, len(latestRows))
	messageIds := make([]string, 0, len(latestRows))
	for _, row := range latestRows {
		latestMessageIdByChannel[row.ChannelId] = row.Id
		messageIds = append(messageIds, row.Id)
	}

	readMessageIds := make(map[string]bool, len(messageIds))
	if len(messageIds) > 0 {
		reads, err := gormdb.Select[chatspg.MessageRead](ctx, databasepg.Query{
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{
				{Key: "message_id", Operator: "IN", Value: messageIds},
				{Key: "account_id", Operator: "=", Value: accountId},
			}}},
		})
		if err != nil {
			return nil, err
		}
		for _, read := range reads {
			readMessageIds[read.MessageId] = true
		}
	}

	for i, channel := range channels {
		latestId, hasMessages := latestMessageIdByChannel[channel.Id]
		channels[i].Unread = hasMessages && !readMessageIds[latestId]
	}

	return channels, nil
}
