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

// SelectParentChannelMessages handles GET
// /private/chats/parent-channels/messages?channel_id= -- the org-less
// counterpart of SelectMessages, authorized via chatspg.ChannelMember
// instead of organization_id (see parentchannel.select.go's doc comment).
func SelectParentChannelMessages(w http.ResponseWriter, r *http.Request) {
	var (
		res chatspg.MessageSelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.SelectParentChannelMessages")
	defer logger.Exit(ctx, "chatsgt.SelectParentChannelMessages", &res, &err)

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	channelId := r.URL.Query().Get("channel_id")
	if channelId == "" {
		http.Error(w, "Missing 'channel_id' query param", http.StatusBadRequest)
		return
	}

	if !requireMember(w, r, channelId, accountId) {
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
