package nfzgt

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Johncode248/Rodzic-Chat/logger"
	nfzpg "github.com/Johncode248/Rodzic-Chat/project/common/nfz"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/nfz"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
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
	if v, e := strconv.ParseFloat(q.Get("lat"), 64); e == nil {
		req.Latitude = v
	}
	if v, e := strconv.ParseFloat(q.Get("lon"), 64); e == nil {
		req.Longitude = v
	}
	if v, e := strconv.Atoi(q.Get("radius_km")); e == nil {
		req.RadiusKm = v
	}
	req.Sort = q.Get("sort")
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

// Specialties handles GET /private/nfz/specialties -- the curated tiles of the start screen.
func Specialties(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(nfzpg.SpecialtiesResponse{Specialties: nfz.Specialties})
}

func viewerId(w http.ResponseWriter, r *http.Request) (string, bool) {
	id, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || id == "" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return "", false
	}
	return id, true
}

// SavedList handles GET /private/nfz/saved -- the parent's "Moje placówki".
func SavedList(w http.ResponseWriter, r *http.Request) {
	id, ok := viewerId(w, r)
	if !ok {
		return
	}
	res, err := nfz.ListSaved(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(res)
}

// SavedAdd handles POST /private/nfz/saved with a SavedPlace body.
func SavedAdd(w http.ResponseWriter, r *http.Request) {
	id, ok := viewerId(w, r)
	if !ok {
		return
	}
	var p nfzpg.SavedPlace
	if json.NewDecoder(r.Body).Decode(&p) != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}
	if err := nfz.AddSaved(r.Context(), id, p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// SavedRemove handles DELETE /private/nfz/saved?queue_id= .
func SavedRemove(w http.ResponseWriter, r *http.Request) {
	id, ok := viewerId(w, r)
	if !ok {
		return
	}
	if err := nfz.RemoveSaved(r.Context(), id, r.URL.Query().Get("queue_id")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
