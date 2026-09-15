package caregiversgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	caregiverspg "github.com/Johncode248/Rodzic-Chat/project/common/caregivers"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/caregivers"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Upsert handles POST /private/caregivers -- creates or updates the
// caller's own nanny/tutor profile. Any account type may call this (a
// parent can also offer these services) -- it's the caller's own
// account_id that owns the row, never something the client can set.
func Upsert(w http.ResponseWriter, r *http.Request) {
	var (
		req caregiverspg.UpsertRequest
		res caregiverspg.Caregiver
		err error
	)

	ctx := logger.Enter(r.Context(), "caregiversgt.Upsert", r.Body)
	defer logger.Exit(ctx, "caregiversgt.Upsert", &res, &err)

	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	res, err = caregivers.Upsert(ctx, accountId, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
