package chatsgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	featurespg "github.com/Johncode248/Rodzic-Chat/project/common/features"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/features"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// pollNotificationBody overrides the generic "sender: content" push body
// message.create.go's notifyNewMessage would otherwise build -- a poll's
// Message.Content is empty, so that path would send a blank-looking push.
const pollNotificationBody = "utworzył(a) ankietę"

// CreatePoll creates a message+poll+options together. Same channel lookup /
// FeatureAnnouncements sub-check as CreateMessage, since a poll is a
// message.
func CreatePoll(w http.ResponseWriter, r *http.Request) {
	var (
		request chatspg.PollCreateRequest
		res     chatspg.PollCreateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.CreatePoll", r.Body)
	defer logger.Exit(ctx, "chatsgt.CreatePoll", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

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

	accountType, err := credentials.GetAccountTypeFromContext(ctx)
	if err != nil {
		http.Error(w, "Error getting accountType from context", http.StatusForbidden)
		return
	}

	channelResult, err := chats.SelectChannels(ctx, chatspg.ChannelSelectArgument{
		Query: databasepg.Query{
			Limit:  1,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: request.ChannelId}}}},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(channelResult.Channels) == 0 || channelResult.Channels[0].OrganizationId != organizationId {
		http.Error(w, chats.ErrChannelNotInOrganization.Error(), http.StatusForbidden)
		return
	}

	if channelResult.Channels[0].Type == chatspg.ChannelTypeAnnouncement {
		enabled, err := features.IsEnabled(ctx, organizationId, featurespg.FeatureAnnouncements)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !enabled {
			http.Error(w, "feature not enabled for this organization", http.StatusForbidden)
			return
		}
	}

	res, err = chats.CreatePoll(ctx, chatspg.PollCreateArgument{
		ChannelId:      request.ChannelId,
		Question:       request.Question,
		Options:        request.Options,
		AllowMultiple:  request.AllowMultiple,
		ClosesAt:       request.ClosesAt,
		OrganizationId: organizationId,
		SenderId:       accountId,
	}, accountType)
	if err != nil {
		status := http.StatusForbidden
		if err == chats.ErrRateLimited {
			status = http.StatusTooManyRequests
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)

	go notifyNewMessage(channelResult.Channels[0], accountId, res.Message, pollNotificationBody)
}
