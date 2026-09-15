package caregiversgt

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Johncode248/Rodzic-Chat/logger"
	caregiverspg "github.com/Johncode248/Rodzic-Chat/project/common/caregivers"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/caregivers"
)

// List handles GET /private/caregivers?type=&query= -- browse nannies
// (type=1) or tutors (type=2), any authenticated account.
func List(w http.ResponseWriter, r *http.Request) {
	var (
		res caregiverspg.ListResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "caregiversgt.List")
	defer logger.Exit(ctx, "caregiversgt.List", &res, &err)

	typeOrdinal, convErr := strconv.Atoi(r.URL.Query().Get("type"))
	if convErr != nil {
		http.Error(w, "Missing or invalid 'type' query param", http.StatusBadRequest)
		return
	}

	res, err = caregivers.List(ctx, caregiverspg.CaregiverTypeFromOrdinal(typeOrdinal), r.URL.Query().Get("query"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
