package parentinvitesgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	parentinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/parentinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentinvites"
	"github.com/gorilla/mux"
)

// PublicInfo handles GET /public/parent-invites/{token} -- unauthenticated
// by design, same as registrationlinksgt.PublicInfo.
func PublicInfo(w http.ResponseWriter, r *http.Request) {
	var (
		res parentinvitespg.PublicInfoResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "parentinvitesgt.PublicInfo")
	defer logger.Exit(ctx, "parentinvitesgt.PublicInfo", &res, &err)

	token := mux.Vars(r)["token"]

	res, err = parentinvites.GetPublicInfo(ctx, token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}

// PublicRegister handles POST /public/parent-invites/{token}/register --
// accepting the invite with no existing account.
func PublicRegister(w http.ResponseWriter, r *http.Request) {
	var (
		req parentinvitespg.RegisterRequest
		res interface{}
		err error
	)

	ctx := logger.Enter(r.Context(), "parentinvitesgt.PublicRegister", r.Body)
	defer logger.Exit(ctx, "parentinvitesgt.PublicRegister", &res, &err)

	token := mux.Vars(r)["token"]

	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	result, err := parentinvites.Register(ctx, token, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res = result.Inherit()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
