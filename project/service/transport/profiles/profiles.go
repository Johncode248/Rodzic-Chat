package profilesgt

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	profilespg "github.com/Johncode248/Rodzic-Chat/project/common/profiles"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/profiles"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

func fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, profiles.ErrKindExists):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, profiles.ErrBadKind), errors.Is(err, profiles.ErrMissing):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, profiles.ErrNotYourGroup):
		http.Error(w, err.Error(), http.StatusForbidden)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// List handles GET /private/profiles.
func List(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Enter(r.Context(), "profilesgt.List")
	var err error
	defer logger.Exit(ctx, "profilesgt.List", &err)
	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	res, err := profiles.List(ctx, accountId)
	if err != nil {
		fail(w, err)
		return
	}
	json.NewEncoder(w).Encode(res)
}

// Add handles POST /private/profiles.
func Add(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Enter(r.Context(), "profilesgt.Add")
	var err error
	defer logger.Exit(ctx, "profilesgt.Add", &err)
	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var req profilespg.AddRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}
	res, err := profiles.Add(ctx, accountId, req)
	if err != nil {
		fail(w, err)
		return
	}
	json.NewEncoder(w).Encode(res)
}

// Switch handles POST /private/profiles/switch.
func Switch(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Enter(r.Context(), "profilesgt.Switch")
	var err error
	defer logger.Exit(ctx, "profilesgt.Switch", &err)
	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var req profilespg.SwitchRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}
	res, err := profiles.Switch(ctx, accountId, req.AccountId)
	if err != nil {
		fail(w, err)
		return
	}
	json.NewEncoder(w).Encode(res)
}
