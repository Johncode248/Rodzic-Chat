package teacherinvitesgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	teacherinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/teacherinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/teacherinvites"
	"github.com/gorilla/mux"
)

// PublicRegister is unauthenticated — completing it is how a teacher gets
// their account. Every validation error (expired/used invite, email
// already taken, ...) is returned as a plain 4xx with a human-readable
// message, since this is rendered straight into a public web form.
func PublicRegister(w http.ResponseWriter, r *http.Request) {
	var (
		request teacherinvitespg.PublicRegisterRequest
		res     teacherinvitespg.PublicRegisterResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "teacherinvitesgt.PublicRegister", r.Body)
	defer logger.Exit(ctx, "teacherinvitesgt.PublicRegister", &res, &err)

	token := mux.Vars(r)["token"]
	if token == "" {
		http.Error(w, "Missing token", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	res, err = teacherinvites.PublicRegister(ctx, teacherinvitespg.PublicRegisterArgument{
		Token:    token,
		Name:     request.Name,
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
