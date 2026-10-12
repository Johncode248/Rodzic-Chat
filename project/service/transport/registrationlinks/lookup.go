package registrationlinksgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	registrationlinkspg "github.com/Johncode248/Rodzic-Chat/project/common/registrationlinks"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/registrationlinks"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Lookup handles POST /private/registration-invites/lookup -- any
// authenticated parent looks up a PIN before committing to Confirm. Rate
// limited (see module.go) since a PIN is brute-forceable in a bounded
// number of guesses otherwise. Does not consume the invite.
func Lookup(w http.ResponseWriter, r *http.Request) {
	var (
		request registrationlinkspg.LookupRequest
		res     registrationlinkspg.LookupResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "registrationlinksgt.Lookup", r.Body)
	defer logger.Exit(ctx, "registrationlinksgt.Lookup", &res, &err)

	// Caller must be authenticated -- this is enforced by AuthMiddleware
	// on the whole /private subrouter, but GetUserIdFromContext also
	// confirms the context actually carries one before proceeding.
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

	res, err = registrationlinks.Lookup(ctx, request.Pin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
