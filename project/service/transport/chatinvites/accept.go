package chatinvitesgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	chatinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/chatinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chatinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Accept handles POST /private/chat-invites/accept -- the other parent,
// already signed in, redeems the PIN to join the inviter's chat channel.
// Rate limited (see module.go).
func Accept(w http.ResponseWriter, r *http.Request) {
	var (
		request chatinvitespg.AcceptRequest
		res     chatinvitespg.AcceptResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "chatinvitesgt.Accept", r.Body)
	defer logger.Exit(ctx, "chatinvitesgt.Accept", &res, &err)

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}
	if request.Pin == "" {
		http.Error(w, "Missing 'pin'", http.StatusBadRequest)
		return
	}

	res, err = chatinvites.Accept(ctx, request.Pin, accountId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(res)
}
