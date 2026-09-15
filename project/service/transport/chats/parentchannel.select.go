package chatsgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// SelectParentChannels handles GET /private/chats/parent-channels -- unlike
// /private/chats/channels (organization-scoped, gated on FeatureChat),
// this lists the caller's org-less, invite-based 1:1 chats (see
// project/common/chatinvites): whatever channel_id rows they have a
// chatspg.ChannelMember row for. Deliberately NOT behind FeatureChat or a
// placówka check -- a parent with no organization at all still has this.
func SelectParentChannels(w http.ResponseWriter, r *http.Request) {
	var (
		res chatspg.ChannelSelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.SelectParentChannels")
	defer logger.Exit(ctx, "chatsgt.SelectParentChannels", &res, &err)

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	memberships, err := gormdb.Select[chatspg.ChannelMember](ctx, databasepg.Query{
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "account_id", Operator: "=", Value: accountId}}},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(memberships) == 0 {
		json.NewEncoder(w).Encode(chatspg.ChannelSelectResponse{Channels: []chatspg.Channel{}})
		return
	}

	channelIds := make([]string, 0, len(memberships))
	for _, m := range memberships {
		channelIds = append(channelIds, m.ChannelId)
	}

	res, err = chats.SelectChannels(ctx, chatspg.ChannelSelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "id", Operator: "IN", Value: channelIds}}},
			},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Channels, err = chats.AttachUnreadFlags(ctx, res.Channels, accountId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
