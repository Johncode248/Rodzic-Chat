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

// DeleteReaction removes the caller's own reaction identified by
// ?message_id=.
func DeleteReaction(w http.ResponseWriter, r *http.Request) {
	var (
		res chatspg.MessageReactionDeleteResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.DeleteReaction")
	defer logger.Exit(ctx, "chatsgt.DeleteReaction", &res, &err)

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	messageId := r.URL.Query().Get("message_id")
	if messageId == "" {
		http.Error(w, "Missing 'message_id' query param", http.StatusBadRequest)
		return
	}

	res, err = chats.DeleteReaction(ctx, chatspg.MessageReactionDeleteArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{
					{Key: "message_id", Operator: "=", Value: messageId},
					{Key: "account_id", Operator: "=", Value: accountId},
				}},
			},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
