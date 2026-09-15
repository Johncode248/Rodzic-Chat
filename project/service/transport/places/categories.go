package placesgt

import (
	"encoding/json"
	"net/http"

	placespg "github.com/Johncode248/Rodzic-Chat/project/common/places"
)

// Categories handles GET /private/places/categories -- the grouped
// picker list a client renders once and reuses for both filtering the
// list and choosing a category when recommending a new place.
func Categories(w http.ResponseWriter, r *http.Request) {
	res := placespg.CategoriesResponse{Groups: placespg.CategoryGroups()}
	json.NewEncoder(w).Encode(res)
}
