package nfzgt

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Johncode248/Rodzic-Chat/logger"
	nfzpg "github.com/Johncode248/Rodzic-Chat/project/common/nfz"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/nfz"
)

// ProvinceForLocation handles GET /private/nfz/province?lat=&lon= -- preselects a sensible
// default województwo for the parent's NFZ search from their device location (see
// nfz.ProvinceForLocation's doc comment for why "nearest capital" is good enough here); the
// client always lets the parent override it manually afterward.
func ProvinceForLocation(w http.ResponseWriter, r *http.Request) {
	var (
		res nfzpg.ProvinceForLocationResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "nfzgt.ProvinceForLocation", r.URL.RawQuery)
	defer logger.Exit(ctx, "nfzgt.ProvinceForLocation", &res, &err)

	lat, latErr := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lon, lonErr := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)
	if latErr != nil || lonErr != nil {
		http.Error(w, "lat and lon are required numeric query params", http.StatusBadRequest)
		return
	}

	code, name := nfz.ProvinceForLocation(lat, lon)
	res = nfzpg.ProvinceForLocationResponse{Province: nfzpg.Province{Code: code, Name: name}}
	json.NewEncoder(w).Encode(res)
}
