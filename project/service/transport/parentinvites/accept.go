package parentinvitesgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	parentinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/parentinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
	"github.com/gorilla/mux"
)

// Accept handles POST /private/parent-invites/{token}/accept -- for a
// caller who already has an account.
func Accept(w http.ResponseWriter, r *http.Request) {
	var (
		res parentinvitespg.AcceptResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "parentinvitesgt.Accept")
	defer logger.Exit(ctx, "parentinvitesgt.Accept", &res, &err)

	token := mux.Vars(r)["token"]

	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	res, err = parentinvites.Accept(ctx, token, accountId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(res)
}
