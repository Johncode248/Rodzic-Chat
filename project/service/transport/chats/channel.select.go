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

// SelectChannels lists channels in the caller's organization.
func SelectChannels(w http.ResponseWriter, r *http.Request) {
	var (
		res chatspg.ChannelSelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.SelectChannels")
	defer logger.Exit(ctx, "chatsgt.SelectChannels", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(ctx)
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	query := databasepg.Query{Groups: []databasepg.Group{{}}}
	query = query.AddOrganizationAuth(organizationId)

	res, err = chats.SelectChannels(ctx, chatspg.ChannelSelectArgument{Query: query})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err == nil && accountId != "" {
		res.Channels, err = chats.AttachUnreadFlags(ctx, res.Channels, accountId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	json.NewEncoder(w).Encode(res)
}
