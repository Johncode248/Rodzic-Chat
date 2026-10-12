package teacherinvitesgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	teacherinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/teacherinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/teacherinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Lookup handles POST /private/teacher-invites/lookup -- any authenticated
// caller looks up a PIN before committing to Confirm. Rate limited (see
// module.go). Does not consume the invite.
func Lookup(w http.ResponseWriter, r *http.Request) {
	var (
		request teacherinvitespg.LookupRequest
		res     teacherinvitespg.LookupResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "teacherinvitesgt.Lookup", r.Body)
	defer logger.Exit(ctx, "teacherinvitesgt.Lookup", &res, &err)

	if _, err = credentials.GetUserIdFromContext(r.Context()); err != nil {
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

	res, err = teacherinvites.Lookup(ctx, request.Pin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
