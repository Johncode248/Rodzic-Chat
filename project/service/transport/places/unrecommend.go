package placesgt

import (
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/places"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Unrecommend handles DELETE /private/places/recommend?place_id= --
// retracts the caller's own recommendation from a place.
func Unrecommend(w http.ResponseWriter, r *http.Request) {
	var err error

	ctx := logger.Enter(r.Context(), "placesgt.Unrecommend")
	defer logger.Exit(ctx, "placesgt.Unrecommend", &err)

	placeId := r.URL.Query().Get("place_id")
	if placeId == "" {
		http.Error(w, "place_id is required", http.StatusBadRequest)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	if err = places.Unrecommend(ctx, accountId, placeId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
