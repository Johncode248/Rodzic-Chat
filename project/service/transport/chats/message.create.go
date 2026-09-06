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

// CreateMessage posts a plain text message. All of /private/chats is
// already gated on FeatureChat at the route level (see module.go); this
// handler additionally checks FeatureAnnouncements when the target channel
// is an Announcement channel, since that sub-feature can be toggled off
// independently even while chat overall stays on. Permission (who may post
// into this channel type) and the anti-spam rate limit are enforced inside
// chats.CreateMessage.
func CreateMessage(w http.ResponseWriter, r *http.Request) {
	var (
		request chatspg.MessageCreateRequest
		res     chatspg.MessageCreateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.CreateMessage", r.Body)
	defer logger.Exit(ctx, "chatsgt.CreateMessage", &res, &err)

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

	res, err = chats.CreateMessage(ctx, chatspg.MessageCreateArgument{
		ChannelId:      request.ChannelId,
		OrganizationId: organizationId,
		SenderId:       accountId,
		Type:           chatspg.MessageTypeText,
		Content:        request.Content,
		AttachmentUrl:  request.AttachmentUrl,
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
}
