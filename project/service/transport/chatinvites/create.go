package chatinvitesgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	chatinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/chatinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chatinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Create handles POST /private/chat-invites -- a parent generates a
// shareable link/token for another parent to start a private 1:1 chat,
// even with no shared placówka (Parent-only, enforced by RequireAccountType
// in module.go).
func Create(w http.ResponseWriter, r *http.Request) {
	var (
		res chatinvitespg.CreateResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "chatinvitesgt.Create")
	defer logger.Exit(ctx, "chatinvitesgt.Create", &res, &err)

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	invite, channel, plaintextPin, err := chatinvites.Create(ctx, accountId)
	if err != nil {
		status := http.StatusInternalServerError
		if err == chatinvites.ErrChannelLimitReached {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}

	res = chatinvitespg.CreateResponse{
		Pin:       plaintextPin,
		ExpiresAt: invite.ExpiresAt.Format("2006-01-02"),
		Channel:   channel,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
