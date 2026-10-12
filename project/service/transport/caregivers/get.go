package caregiversgt

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Johncode248/Rodzic-Chat/logger"
	caregiverspg "github.com/Johncode248/Rodzic-Chat/project/common/caregivers"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/caregivers"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// GetMine handles GET /private/caregivers/mine?type= -- lets the client
// know whether the caller already has a profile for that type (prefill
// the edit form) or should be offered "Dodaj swój profil" instead. Empty
// Caregiver{} + 404 means "none yet", not an error to surface loudly.
func GetMine(w http.ResponseWriter, r *http.Request) {
	var (
		res caregiverspg.Caregiver
		err error
	)

	ctx := logger.Enter(r.Context(), "caregiversgt.GetMine")
	defer logger.Exit(ctx, "caregiversgt.GetMine", &res, &err)

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	typeOrdinal, convErr := strconv.Atoi(r.URL.Query().Get("type"))
	if convErr != nil {
		http.Error(w, "Missing or invalid 'type' query param", http.StatusBadRequest)
		return
	}

	res, err = caregivers.GetMine(ctx, accountId, caregiverspg.CaregiverTypeFromOrdinal(typeOrdinal))
	if err != nil {
		if err == caregivers.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}

// GetDetail handles GET /private/caregivers/detail?id=.
func GetDetail(w http.ResponseWriter, r *http.Request) {
	var (
		res caregiverspg.Detail
		err error
	)

	ctx := logger.Enter(r.Context(), "caregiversgt.GetDetail")
	defer logger.Exit(ctx, "caregiversgt.GetDetail", &res, &err)

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing 'id' query param", http.StatusBadRequest)
		return
	}

	res, err = caregivers.GetDetail(ctx, id, accountId)
	if err != nil {
		if err == caregivers.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}

// ListMine handles GET /private/caregivers/my -- every profile the caller owns with its review status.
func ListMine(w http.ResponseWriter, r *http.Request) {
	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}
	res, err := caregivers.ListMine(r.Context(), accountId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(res)
}
