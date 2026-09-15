package chatinvitesgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	chatinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/chatinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chatinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
	"github.com/gorilla/mux"
)

// Accept handles POST /private/chat-invites/{token}/accept -- the other
// parent, already signed in, redeems the code/link to join the inviter's
// chat channel.
func Accept(w http.ResponseWriter, r *http.Request) {
	var (
		res chatinvitespg.AcceptResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "chatinvitesgt.Accept")
	defer logger.Exit(ctx, "chatinvitesgt.Accept", &res, &err)

	token := mux.Vars(r)["token"]

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	res, err = chatinvites.Accept(ctx, token, accountId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(res)
}
