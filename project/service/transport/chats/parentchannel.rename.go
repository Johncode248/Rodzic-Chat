package chatsgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

type renameParentChannelRequest struct {
	ChannelId string `json:"channel_id"`
	Name      string `json:"name"`
}

// RenameParentChannel handles PATCH /private/chats/parent-channels -- "Zmień nazwę" from a
// channel's settings screen. Admin-only; see chats.RenameChannel's doc comment.
func RenameParentChannel(w http.ResponseWriter, r *http.Request) {
	var (
		res chatspg.Channel
		err error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.RenameParentChannel")
	defer logger.Exit(ctx, "chatsgt.RenameParentChannel", &res, &err)

	callerAccountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || callerAccountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	var req renameParentChannelRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.ChannelId == "" {
		http.Error(w, "Missing 'channel_id'", http.StatusBadRequest)
		return
	}

	res, err = chats.RenameChannel(ctx, req.ChannelId, callerAccountId, req.Name)
	if err != nil {
		status := http.StatusInternalServerError
		switch err {
		case chats.ErrNotChannelAdmin:
			status = http.StatusForbidden
		case chats.ErrChannelNameRequired, chats.ErrChannelNameTooLong:
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}

	json.NewEncoder(w).Encode(res)
}
