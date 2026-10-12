package chatinvitesgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	chatinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/chatinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chatinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Lookup handles POST /private/chat-invites/lookup -- an already
// authenticated caller previews a PIN (inviter's name) before committing
// to Accept. Rate limited (see module.go). Does not consume the invite.
func Lookup(w http.ResponseWriter, r *http.Request) {
	var (
		request chatinvitespg.LookupRequest
		res     chatinvitespg.LookupResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "chatinvitesgt.Lookup", r.Body)
	defer logger.Exit(ctx, "chatinvitesgt.Lookup", &res, &err)

	if _, err = credentials.GetUserIdFromContext(ctx); err != nil {
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

	res, err = chatinvites.Lookup(ctx, request.Pin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
