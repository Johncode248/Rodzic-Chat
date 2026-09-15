package places

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	placespg "github.com/Johncode248/Rodzic-Chat/project/common/places"
	"github.com/Johncode248/Rodzic-Chat/variables"
)

var ErrGooglePlacesNotConfigured = errors.New("Google Places search is not configured yet -- missing GOOGLE_PLACES_API_KEY")

const googleTextSearchURL = "https://maps.googleapis.com/maps/api/place/textsearch/json"

type googleTextSearchResult struct {
	PlaceId          string   `json:"place_id"`
	Name             string   `json:"name"`
	FormattedAddress string   `json:"formatted_address"`
	Types            []string `json:"types"`
	Geometry         struct {
		Location struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"location"`
	} `json:"geometry"`
}

// categoryFromGoogleTypes mirrors categoryFromOSMTag (osm.go) for Google's own place "types"
// array -- https://developers.google.com/maps/documentation/places/web-service/place-types.
// Checked in order since a place carries several types; the first recognized one wins.
func categoryFromGoogleTypes(types []string) string {
	mapping := map[string]string{
		"pharmacy":        "apteka_calodobowa",
		"doctor":          "pediatra",
		"dentist":         "stomatolog_dzieciecy",
		"hospital":        "szpital_pediatryczny",
		"school":          "szkola_podstawowa",
		"library":         "biblioteka",
		"veterinary_care": "inne",
		"zoo":             "zoo_gospodarstwo",
		"museum":          "muzeum",
		"amusement_park":  "urodziny_dla_dzieci",
	}
	for _, t := range types {
		if category, ok := mapping[t]; ok {
			return category
		}
	}
	return ""
}

type googleTextSearchResponse struct {
	Results []googleTextSearchResult `json:"results"`
	Status  string                   `json:"status"`
	Error   string                   `json:"error_message"`
}

// SearchGoogle is the Pro-tier alternative to Search (OpenStreetMap) -- much better coverage of
// small local businesses (individual doctors, specialists, tutors) that volunteer-mapped OSM
// often lacks, at the cost of actually costing money per call. Gated to Pro/Max plans at the
// transport layer (see transport/places/search.go), not here.
func SearchGoogle(ctx context.Context, query string) (candidates []placespg.SearchCandidate, err error) {
	if variables.GooglePlacesApiKey == "" {
		return nil, ErrGooglePlacesNotConfigured
	}
	if query == "" {
		return nil, nil
	}

	reqURL := fmt.Sprintf("%s?%s", googleTextSearchURL, url.Values{
		"query":    {query},
		"language": {"pl"},
		"region":   {"pl"},
		"key":      {variables.GooglePlacesApiKey},
	}.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var parsed googleTextSearchResponse
	if err = json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	if parsed.Status != "OK" && parsed.Status != "ZERO_RESULTS" {
		if parsed.Error != "" {
			return nil, fmt.Errorf("places: Google error (%s): %s", parsed.Status, parsed.Error)
		}
		return nil, fmt.Errorf("places: Google error status %s", parsed.Status)
	}

	for _, r := range parsed.Results {
		candidates = append(candidates, placespg.SearchCandidate{
			Source:            placespg.SourceGoogle,
			ExternalRef:       r.PlaceId,
			Name:              r.Name,
			Address:           r.FormattedAddress,
			Latitude:          r.Geometry.Location.Lat,
			Longitude:         r.Geometry.Location.Lng,
			SuggestedCategory: categoryFromGoogleTypes(r.Types),
		})
	}

	return candidates, nil
}
