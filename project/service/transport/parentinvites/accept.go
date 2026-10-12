package parentinvitesgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	parentinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/parentinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Accept handles POST /private/parent-invites/accept -- an
// already-authenticated caller redeems a PIN to link to the inviter's
// children. Rate limited (see module.go).
func Accept(w http.ResponseWriter, r *http.Request) {
	var (
		request parentinvitespg.AcceptRequest
		res     parentinvitespg.AcceptResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "parentinvitesgt.Accept", r.Body)
	defer logger.Exit(ctx, "parentinvitesgt.Accept", &res, &err)

	accountId, err := credentials.GetUserIdFromContext(r.Context())
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

	res, err = parentinvites.Accept(ctx, request.Pin, accountId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(res)
}
