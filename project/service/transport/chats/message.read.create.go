package chatsgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// MarkRead records (or refreshes) the caller having read a message --
// idempotent upsert on (message_id, account_id).
func MarkRead(w http.ResponseWriter, r *http.Request) {
	var (
		request chatspg.MessageReadCreateRequest
		res     chatspg.MessageReadCreateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.MarkRead", r.Body)
	defer logger.Exit(ctx, "chatsgt.MarkRead", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	res, err = chats.MarkRead(ctx, chatspg.MessageReadCreateArgument{
		MessageId: request.MessageId,
		AccountId: accountId,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
