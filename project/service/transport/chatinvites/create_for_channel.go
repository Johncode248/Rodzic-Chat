package chatinvitesgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	chatinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/chatinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chatinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

type createForChannelRequest struct {
	ChannelId string `json:"channel_id,omitempty"`
}

// CreateForChannel handles POST /private/chat-invites/for-channel --
// "Zaproś kolejną osobę" from within an existing channel's settings.
// Admin-only (see chats.IsChannelAdmin).
func CreateForChannel(w http.ResponseWriter, r *http.Request) {
	var (
		req createForChannelRequest
		res chatinvitespg.CreateResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "chatinvitesgt.CreateForChannel", r.Body)
	defer logger.Exit(ctx, "chatinvitesgt.CreateForChannel", &res, &err)

	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}
	if req.ChannelId == "" {
		http.Error(w, "Missing 'channel_id'", http.StatusBadRequest)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	invite, plaintextPin, err := chatinvites.CreateForChannel(ctx, req.ChannelId, accountId)
	if err != nil {
		status := http.StatusInternalServerError
		if err == chats.ErrNotChannelAdmin {
			status = http.StatusForbidden
		}
		http.Error(w, err.Error(), status)
		return
	}

	res = chatinvitespg.CreateResponse{
		Pin:       plaintextPin,
		ExpiresAt: invite.ExpiresAt.Format("2006-01-02"),
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
