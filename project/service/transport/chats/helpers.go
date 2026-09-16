package chatsgt

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
)

// maxMessageListLimit bounds ?limit= on the message-listing endpoints --
// without it, ?limit=999999999 forces the backend to load (and JSON-encode)
// an entire channel's history in one request.
const maxMessageListLimit = 200

// messageInOrganization verifies that messageId belongs to organizationId
// (Message carries its own OrganizationId column) -- the multi-tenancy
// check every reaction/read-receipt route that takes a bare ?message_id=
// must perform, mirroring the channel_id check already done in
// CreateMessage/SelectMessages/CreatePoll.
func messageInOrganization(ctx context.Context, messageId, organizationId string) (bool, error) {
	if messageId == "" || organizationId == "" {
		return false, nil
	}

	result, err := chats.SelectMessages(ctx, chatspg.MessageSelectArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{
					{Key: "id", Operator: "=", Value: messageId},
					{Key: "organization_id", Operator: "=", Value: organizationId},
				}},
			},
		},
	})
	if err != nil {
		return false, err
	}

	return len(result.Messages) > 0, nil
}

// channelForMessage resolves the channel a message belongs to, so callers
// (reactions/reads/polls) can run chats.CheckReadPermission the same way
// SelectMessages does -- messageInOrganization alone only proves the
// message is in the caller's organization, not that the caller is actually
// a member of that message's (possibly group-restricted) channel.
func channelForMessage(ctx context.Context, messageId string) (chatspg.Channel, bool, error) {
	messageResult, err := chats.SelectMessages(ctx, chatspg.MessageSelectArgument{
		Query: databasepg.Query{
			Limit:  1,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: messageId}}}},
		},
	})
	if err != nil {
		return chatspg.Channel{}, false, err
	}
	if len(messageResult.Messages) == 0 {
		return chatspg.Channel{}, false, nil
	}

	channelResult, err := chats.SelectChannels(ctx, chatspg.ChannelSelectArgument{
		Query: databasepg.Query{
			Limit:  1,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: messageResult.Messages[0].ChannelId}}}},
		},
	})
	if err != nil {
		return chatspg.Channel{}, false, err
	}
	if len(channelResult.Channels) == 0 {
		return chatspg.Channel{}, false, nil
	}

	return channelResult.Channels[0], true, nil
}

// pollInOrganization verifies that pollId's owning message belongs to
// organizationId -- Poll has no organization_id of its own, so this goes
// through Poll.MessageId, same as SelectPoll already does inline.
func pollInOrganization(ctx context.Context, pollId, organizationId string) (bool, error) {
	if pollId == "" || organizationId == "" {
		return false, nil
	}

	pollResult, err := chats.SelectPolls(ctx, chatspg.PollSelectArgument{
		Query: databasepg.Query{
			Limit:  1,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: pollId}}}},
		},
	})
	if err != nil {
		return false, err
	}
	if len(pollResult.Polls) == 0 {
		return false, nil
	}

	return messageInOrganization(ctx, pollResult.Polls[0].MessageId, organizationId)
}
