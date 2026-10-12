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

// SelectReactions lists every reaction on ?message_id= -- reactions are
// public within the org (like a "like" list), so no per-caller filtering.
func SelectReactions(w http.ResponseWriter, r *http.Request) {
	var (
		res chatspg.MessageReactionSelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.SelectReactions")
	defer logger.Exit(ctx, "chatsgt.SelectReactions", &res, &err)

	messageId := r.URL.Query().Get("message_id")
	if messageId == "" {
		http.Error(w, "Missing 'message_id' query param", http.StatusBadRequest)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	// Note: organizationId may legitimately be "" here -- see
	// messageAccessible.
	organizationId, err := credentials.GetOrganizationIdFromContext(ctx)
	if err != nil {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	accountType, err := credentials.GetAccountTypeFromContext(ctx)
	if err != nil {
		http.Error(w, "Error getting accountType from context", http.StatusForbidden)
		return
	}
	channel, found, allowed, err := messageAccessible(ctx, messageId, organizationId, accountId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !found {
		http.Error(w, "message not found", http.StatusNotFound)
		return
	}
	if !allowed {
		http.Error(w, chats.ErrChannelNotInOrganization.Error(), http.StatusForbidden)
		return
	}
	if err = chats.CheckReadPermission(ctx, channel, accountId, accountType); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	res, err = chats.SelectReactions(ctx, chatspg.MessageReactionSelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "message_id", Operator: "=", Value: messageId}}},
			},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
