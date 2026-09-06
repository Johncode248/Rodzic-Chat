package chatsgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
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
