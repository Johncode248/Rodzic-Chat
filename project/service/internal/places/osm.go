package places

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	placespg "github.com/Johncode248/Rodzic-Chat/project/common/places"
)

const nominatimSearchURL = "https://nominatim.openstreetmap.org/search"

// nominatimResult is the subset of Nominatim's response we use. osm_type +
// osm_id together form the stable external reference we de-duplicate
// against later (see recommend.go) -- Nominatim's own "place_id" is NOT
// stable across their index rebuilds, unlike osm_type/osm_id.
//
// Class/Type are OSM's own tagging ("amenity"/"pharmacy", "tourism"/"museum",
// ...) -- used only to guess one of our own Category keys (see
// categoryFromOSMTag) so recommending a search result can be a single tap
// instead of forcing a category picker every time. ExtraTags carries
// whatever else OSM happens to have (phone, website, opening_hours, ...),
// entirely optional and inconsistent between places by nature of
// volunteer-mapped data.
type nominatimResult struct {
	OSMType     string            `json:"osm_type"`
	OSMId       int64             `json:"osm_id"`
	Class       string            `json:"class"`
	Type        string            `json:"type"`
	DisplayName string            `json:"display_name"`
	Lat         string            `json:"lat"`
	Lon         string            `json:"lon"`
	ExtraTags   map[string]string `json:"extratags"`
}

func firstNonEmpty(tags map[string]string, keys ...string) string {
	for _, key := range keys {
		if v, ok := tags[key]; ok && v != "" {
			return v
		}
	}
	return ""
}

// categoryFromOSMTag guesses one of our own Category keys (see
// project/common/places.CategoryGroups) from OSM's class/type tag pair.
// Deliberately covers only the common, unambiguous cases -- anything else
// returns "" and the client falls back to asking the parent to pick one,
// same as a manually-added place.
func categoryFromOSMTag(class string, osmType string) string {
	switch class + "/" + osmType {
	case "amenity/pharmacy":
		return "apteka_calodobowa"
	case "amenity/doctors", "amenity/clinic":
		return "pediatra"
	case "amenity/dentist":
		return "stomatolog_dzieciecy"
	case "amenity/hospital":
		return "szpital_pediatryczny"
	case "amenity/kindergarten":
		return "przedszkole"
	case "amenity/school":
		return "szkola_podstawowa"
	case "amenity/library":
		return "biblioteka"
	case "amenity/veterinary":
		return "inne"
	case "leisure/playground":
		return "plac_zabaw"
	case "leisure/water_park":
		return "park_trampolin_baseny"
	case "tourism/museum":
		return "muzeum"
	case "tourism/zoo":
		return "zoo_gospodarstwo"
	case "tourism/theme_park":
		return "urodziny_dla_dzieci"
	case "shop/toys":
		return "zabawki"
	case "shop/clothes", "shop/babygoods":
		return "odziez_dziecieca"
	default:
		return ""
	}
}

// Search proxies OpenStreetMap's free Nominatim geocoder. Nominatim's usage
// policy requires a descriptive User-Agent and caps unauthenticated traffic
// at ~1 request/second -- fine for this app's expected volume, but this
// must never be called in a tight loop.
func Search(ctx context.Context, query string) (candidates []placespg.SearchCandidate, err error) {
	if query == "" {
		return nil, nil
	}

	reqURL := fmt.Sprintf("%s?%s", nominatimSearchURL, url.Values{
		"q":              {query},
		"format":         {"json"},
		"limit":          {"10"},
		"addressdetails": {"0"},
		"extratags":      {"1"},
	}.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "RodzicChat/1.0 (contact: kontakt@rodzicchat.pl)")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var results []nominatimResult
	if err = json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	for _, r := range results {
		lat, latErr := strconv.ParseFloat(r.Lat, 64)
		lon, lonErr := strconv.ParseFloat(r.Lon, 64)
		if latErr != nil || lonErr != nil {
			continue
		}
		// Nominatim's display_name is "Name, street, district, city, ...";
		// splitting off the first component gives a usable short name while
		// the full string still makes a perfectly good address.
		name := r.DisplayName
		if idx := strings.Index(r.DisplayName, ","); idx > 0 {
			name = r.DisplayName[:idx]
		}

		candidates = append(candidates, placespg.SearchCandidate{
			Source:            placespg.SourceOSM,
			ExternalRef:       fmt.Sprintf("%s/%d", r.OSMType, r.OSMId),
			Name:              name,
			Address:           r.DisplayName,
			Latitude:          lat,
			Longitude:         lon,
			Phone:             firstNonEmpty(r.ExtraTags, "phone", "contact:phone"),
			Website:           firstNonEmpty(r.ExtraTags, "website", "contact:website", "url"),
			OpeningHours:      r.ExtraTags["opening_hours"],
			SuggestedCategory: categoryFromOSMTag(r.Class, r.Type),
		})
	}

	return candidates, nil
}
