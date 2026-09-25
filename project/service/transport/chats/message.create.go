package chatsgt

import (
	"context"
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	featurespg "github.com/Johncode248/Rodzic-Chat/project/common/features"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/features"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/push"
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
		Urgent:         request.Urgent,
		ReplyToId:      request.ReplyToId,
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

	// Push notification: best-effort, fire-and-forget on its own context
	// (the request's r.Context() is cancelled the instant this handler
	// returns, which just did) so a slow FCM call can't hang the response
	// and a failed one can never turn a sent message into an error.
	go notifyNewMessage(channelResult.Channels[0], accountId, res.Message, "")
}

// bodyOverride replaces "{sender name}: {message content}" wholesale when
// non-empty -- used by poll.create.go, whose Message.Content is always
// blank (the question lives on the Poll row, not the Message).
func notifyNewMessage(channel chatspg.Channel, senderId string, message chatspg.Message, bodyOverride string) {
	ctx := context.Background()

	recipientIds, err := chats.Recipients(ctx, channel)
	if err != nil {
		return
	}

	targets := make([]string, 0, len(recipientIds))
	for _, id := range recipientIds {
		if id != senderId {
			targets = append(targets, id)
		}
	}
	if len(targets) == 0 {
		return
	}

	senderName := "Ktoś"
	if senderResult, err := accounts.Select(ctx, accountspg.SelectArgument{
		Query: databasepg.Query{
			Limit:  1,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: senderId}}}},
		},
	}); err == nil && len(senderResult.Accounts) > 0 {
		senderName = senderResult.Accounts[0].Name
	}

	title := channel.Name
	body := message.Content
	if channel.Type == chatspg.ChannelTypeDirect {
		title = senderName
	} else {
		body = senderName + ": " + message.Content
	}
	if bodyOverride != "" {
		body = senderName + " " + bodyOverride
	} else if body == "" || body == senderName+": " {
		body = senderName + " wysłał(a) załącznik"
	}
	const maxBodyLen = 160
	if len(body) > maxBodyLen {
		body = body[:maxBodyLen] + "…"
	}

	if message.Urgent {
		title = "PILNE · " + title
	}

	push.SendToAccounts(ctx, targets, title, body, map[string]string{
		"type":       "chat_message",
		"channel_id": channel.Id,
	})
}
