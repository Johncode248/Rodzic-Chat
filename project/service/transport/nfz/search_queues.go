package nfzgt

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Johncode248/Rodzic-Chat/logger"
	nfzpg "github.com/Johncode248/Rodzic-Chat/project/common/nfz"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/nfz"
)

// SearchQueues handles GET /private/nfz/queues -- the manual search form on the Miejsca tab's
// "Opieka zdrowotna" side. [benefit] must be an exact NFZ name (see FindBenefit).
func SearchQueues(w http.ResponseWriter, r *http.Request) {
	var (
		res nfzpg.SearchQueuesResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "nfzgt.SearchQueues", r.URL.RawQuery)
	defer logger.Exit(ctx, "nfzgt.SearchQueues", &res, &err)

	q := r.URL.Query()

	req := nfzpg.SearchQueuesRequest{
		Province:     q.Get("province"),
		Benefit:      q.Get("benefit"),
		Locality:     q.Get("locality"),
		ChildAgeBand: q.Get("child_age_band"),
	}
	if q.Get("case") == "2" {
		req.Case = nfzpg.CaseUrgent
	} else {
		req.Case = nfzpg.CaseStable
	}
	if page, parseErr := strconv.Atoi(q.Get("page")); parseErr == nil {
		req.Page = page
	}
	if limit, parseErr := strconv.Atoi(q.Get("limit")); parseErr == nil {
		req.Limit = limit
	}

	if req.Benefit == "" {
		http.Error(w, "benefit is required -- resolve one via GET /nfz/benefits first", http.StatusBadRequest)
		return
	}

	res, err = nfz.SearchQueues(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	json.NewEncoder(w).Encode(res)
}
