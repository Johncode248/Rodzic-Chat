package chatsgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// SelectPoll fetches the poll+options+vote tally+"my votes" aggregate for
// ?poll_id= (or ?message_id=, for a client that only has the poll-type
// Message and never cached its poll id — e.g. after re-opening a channel's
// history), verifying the poll's owning message belongs to the caller's
// organization.
func SelectPoll(w http.ResponseWriter, r *http.Request) {
	var (
		res chatspg.PollDetailResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.SelectPoll")
	defer logger.Exit(ctx, "chatsgt.SelectPoll", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(ctx)
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	pollId := r.URL.Query().Get("poll_id")
	messageId := r.URL.Query().Get("message_id")
	if pollId == "" && messageId == "" {
		http.Error(w, "Missing 'poll_id' or 'message_id' query param", http.StatusBadRequest)
		return
	}

	lookupKey, lookupValue := "id", pollId
	if pollId == "" {
		lookupKey, lookupValue = "message_id", messageId
	}

	pollResult, err := chats.SelectPolls(ctx, chatspg.PollSelectArgument{
		Query: databasepg.Query{
			Limit:  1,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: lookupKey, Operator: "=", Value: lookupValue}}}},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(pollResult.Polls) == 0 {
		http.Error(w, "poll not found", http.StatusNotFound)
		return
	}
	pollId = pollResult.Polls[0].Id

	messageResult, err := chats.SelectMessages(ctx, chatspg.MessageSelectArgument{
		Query: databasepg.Query{
			Limit:  1,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: pollResult.Polls[0].MessageId}}}},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(messageResult.Messages) == 0 || messageResult.Messages[0].OrganizationId != organizationId {
		http.Error(w, chats.ErrChannelNotInOrganization.Error(), http.StatusForbidden)
		return
	}

	accountType, err := credentials.GetAccountTypeFromContext(ctx)
	if err != nil {
		http.Error(w, "Error getting accountType from context", http.StatusForbidden)
		return
	}
	channelResult, err := chats.SelectChannels(ctx, chatspg.ChannelSelectArgument{
		Query: databasepg.Query{
			Limit:  1,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: messageResult.Messages[0].ChannelId}}}},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(channelResult.Channels) == 0 {
		http.Error(w, "channel not found", http.StatusNotFound)
		return
	}
	if err = chats.CheckReadPermission(ctx, channelResult.Channels[0], accountId, accountType); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	res, err = chats.GetPollDetail(ctx, pollId, accountId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
