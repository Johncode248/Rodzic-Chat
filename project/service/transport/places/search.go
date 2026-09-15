package placesgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	placespg "github.com/Johncode248/Rodzic-Chat/project/common/places"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/places"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Search handles GET /private/places/search?query= -- proxies a live place search for a parent
// to pick a candidate from when adding a place that isn't recommended by anyone yet (see
// placesgt.Recommend). Read-only, never writes anything -- picking a candidate is a separate
// Recommend call.
//
// Every parent goes through places.SearchWithFallback now, which tries Google Places first
// (much better coverage of small local businesses -- individual doctors, specialists -- than
// OpenStreetMap) and transparently falls back to OpenStreetMap once the shared Google budget or
// this parent's own daily cap runs out. No plan gate, no explicit provider choice on the client
// side anymore -- see SearchWithFallback's doc comment.
func Search(w http.ResponseWriter, r *http.Request) {
	var (
		res placespg.SearchResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "placesgt.Search")
	defer logger.Exit(ctx, "placesgt.Search", &res, &err)

	query := r.URL.Query().Get("query")
	if query == "" {
		json.NewEncoder(w).Encode(res)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	candidates, provider, err := places.SearchWithFallback(ctx, accountId, query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res = placespg.SearchResponse{Candidates: candidates, Provider: provider}
	json.NewEncoder(w).Encode(res)
}
