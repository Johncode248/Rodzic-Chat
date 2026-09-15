package placesgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	placespg "github.com/Johncode248/Rodzic-Chat/project/common/places"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/places"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// List handles GET /private/places?category=&query=&sort= -- every place
// already recommended by at least one parent (never OSM data directly),
// filterable by category and a free-text name/address search, sorted by
// recommendation count (default) or name.
func List(w http.ResponseWriter, r *http.Request) {
	var (
		res placespg.ListResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "placesgt.List")
	defer logger.Exit(ctx, "placesgt.List", &res, &err)

	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	list, err := places.List(
		ctx,
		accountId,
		r.URL.Query().Get("category"),
		r.URL.Query().Get("query"),
		r.URL.Query().Get("sort"),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res = placespg.ListResponse{Places: list}
	json.NewEncoder(w).Encode(res)
}
