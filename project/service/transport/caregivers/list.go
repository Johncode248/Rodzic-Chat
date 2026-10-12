package caregiversgt

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Johncode248/Rodzic-Chat/logger"
	caregiverspg "github.com/Johncode248/Rodzic-Chat/project/common/caregivers"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/caregivers"
)

// maxCaregiverListLimit bounds ?limit= the same way the message/channel-listing endpoints do.
const maxCaregiverListLimit = 200

// List handles GET /private/caregivers?type=&query=&limit=&offset= -- browse nannies
// (type=1) or tutors (type=2), any authenticated account. limit/offset are optional; omitting
// both keeps the previous "return everything" behavior for any existing caller.
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

	limit := 0
	if l, convErr := strconv.Atoi(r.URL.Query().Get("limit")); convErr == nil && l > 0 {
		limit = l
		if limit > maxCaregiverListLimit {
			limit = maxCaregiverListLimit
		}
	}
	offset := 0
	if o, convErr := strconv.Atoi(r.URL.Query().Get("offset")); convErr == nil && o > 0 {
		offset = o
	}

	var maxRate int64
	if m, convErr := strconv.ParseInt(r.URL.Query().Get("max_rate"), 10, 64); convErr == nil && m > 0 {
		maxRate = m
	}
	res, err = caregivers.Search(ctx, caregivers.SearchFilter{
		Type: caregiverspg.CaregiverTypeFromOrdinal(typeOrdinal), Query: r.URL.Query().Get("query"),
		City: r.URL.Query().Get("city"), Mode: r.URL.Query().Get("mode"), MaxRate: maxRate, Limit: limit, Offset: offset,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
