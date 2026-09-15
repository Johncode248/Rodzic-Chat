package placesgt

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	placespg "github.com/Johncode248/Rodzic-Chat/project/common/places"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/places"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Recommend handles POST /private/places/recommend -- recommends an
// existing place (place_id) or first creates one (from an OSM search
// candidate or typed by hand) and recommends that. Repeat calls from the
// same parent on the same place update their comment rather than erroring.
func Recommend(w http.ResponseWriter, r *http.Request) {
	var (
		req placespg.RecommendRequest
		res placespg.PlaceWithStats
		err error
	)

	ctx := logger.Enter(r.Context(), "placesgt.Recommend", r.Body)
	defer logger.Exit(ctx, "placesgt.Recommend", &res, &err)

	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	if err = req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	res, err = places.Recommend(ctx, accountId, req)
	if err != nil {
		if errors.Is(err, places.ErrRecommendationLimitReached) {
			http.Error(w, err.Error(), http.StatusPaymentRequired)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
