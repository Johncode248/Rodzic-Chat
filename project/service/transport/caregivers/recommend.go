package caregiversgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	caregiverspg "github.com/Johncode248/Rodzic-Chat/project/common/caregivers"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/caregivers"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Recommend handles POST /private/caregivers/recommendations -- a parent's
// "Polecam" endorsement of a caregiver profile that isn't their own.
func Recommend(w http.ResponseWriter, r *http.Request) {
	var (
		req caregiverspg.RecommendRequest
		res caregiverspg.CaregiverRecommendation
		err error
	)

	ctx := logger.Enter(r.Context(), "caregiversgt.Recommend", r.Body)
	defer logger.Exit(ctx, "caregiversgt.Recommend", &res, &err)

	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	res, err = caregivers.Recommend(ctx, req.CaregiverId, accountId, req.Comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

// Unrecommend handles DELETE /private/caregivers/recommendations?caregiver_id=.
func Unrecommend(w http.ResponseWriter, r *http.Request) {
	var err error

	ctx := logger.Enter(r.Context(), "caregiversgt.Unrecommend")
	defer logger.Exit(ctx, "caregiversgt.Unrecommend", &err)

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	caregiverId := r.URL.Query().Get("caregiver_id")
	if caregiverId == "" {
		http.Error(w, "Missing 'caregiver_id' query param", http.StatusBadRequest)
		return
	}

	if err = caregivers.Unrecommend(ctx, caregiverId, accountId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
