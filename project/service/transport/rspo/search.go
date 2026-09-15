package rspogt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	rspopg "github.com/Johncode248/Rodzic-Chat/project/common/rspo"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/rspo"
)

// Search handles GET /private/rspo/search?query=&category= -- real RSPO (Ministry of Education)
// data for the Wyszukaj hub's education/support categories, instead of a guessed OpenStreetMap
// keyword search. See project/common/rspo's doc comment for exactly what RSPO does and doesn't
// cover.
func Search(w http.ResponseWriter, r *http.Request) {
	var (
		res rspopg.SearchResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "rspogt.Search", r.URL.RawQuery)
	defer logger.Exit(ctx, "rspogt.Search", &res, &err)

	q := r.URL.Query()
	category := rspopg.Category(q.Get("category"))
	if category == "" {
		http.Error(w, "category is required", http.StatusBadRequest)
		return
	}

	institutions, err := rspo.Search(ctx, q.Get("query"), category, 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	res = rspopg.SearchResponse{Institutions: institutions}
	json.NewEncoder(w).Encode(res)
}
