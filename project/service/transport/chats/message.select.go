package chatsgt

import (
	"encoding/json"
	"net/http"
	"strconv"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// SelectMessages lists messages for ?channel_id=, newest-first, paginated
// via ?limit= / ?offset=.
func SelectMessages(w http.ResponseWriter, r *http.Request) {
	var (
		res chatspg.MessageSelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.SelectMessages")
	defer logger.Exit(ctx, "chatsgt.SelectMessages", &res, &err)

	// Note: organizationId may legitimately be "" here (a parent with no
	// organization yet, e.g. one only using an org-less chat-invite
	// channel) -- do not gate on it being non-empty, only use it below to
	// authorize access to an org-scoped channel.
	organizationId, err := credentials.GetOrganizationIdFromContext(ctx)
	if err != nil {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	channelId := r.URL.Query().Get("channel_id")
	if channelId == "" {
		http.Error(w, "Missing 'channel_id' query param", http.StatusBadRequest)
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
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: channelId}}}},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(channelResult.Channels) == 0 {
		http.Error(w, chats.ErrChannelNotInOrganization.Error(), http.StatusForbidden)
		return
	}
	channel := channelResult.Channels[0]

	// Org-scoped channel: same organization match as before. Org-less
	// channel (e.g. a chat-invite-created Direct channel, OrganizationId
	// == ""): there's no organization to match against, so membership in
	// the channel itself is the access control -- see chatinvites.Accept,
	// which is the only way to end up a member of one of these.
	if channel.OrganizationId != "" {
		if channel.OrganizationId != organizationId {
			http.Error(w, chats.ErrChannelNotInOrganization.Error(), http.StatusForbidden)
			return
		}
	} else {
		isMember, err := chats.IsChannelMember(ctx, channelId, accountId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !isMember {
			http.Error(w, chats.ErrChannelNotInOrganization.Error(), http.StatusForbidden)
			return
		}
	}

	if err = chats.CheckReadPermission(ctx, channel, accountId, accountType); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	limit := 50
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 {
		limit = l
		if limit > maxMessageListLimit {
			limit = maxMessageListLimit
		}
	}
	offset := 0
	if o, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && o > 0 {
		offset = o
	}

	res, err = chats.SelectMessages(ctx, chatspg.MessageSelectArgument{
		Query: databasepg.Query{
			Limit:  limit,
			Offset: offset,
			OrderBy: databasepg.Order{
				Key:       "create_time",
				Direction: databasepg.DirectionDescending,
			},
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "channel_id", Operator: "=", Value: channelId}}},
			},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Messages, err = chats.AttachSenderNames(ctx, res.Messages)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
