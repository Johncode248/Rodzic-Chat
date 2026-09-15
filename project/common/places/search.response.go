package placespg

// SearchCandidate is one OpenStreetMap Nominatim search hit -- not yet a
// Place in our own database. The client shows these when a parent is
// adding a place that isn't already recommended by anyone; picking one and
// submitting RecommendRequest with ExternalRef+Source="osm" set is what
// actually creates the Place row (see
// project/service/internal/places/recommend.go).
type SearchCandidate struct {
	Source      string  `json:"source,omitempty"`
	ExternalRef string  `json:"external_ref,omitempty"`
	Name        string  `json:"name,omitempty"`
	Address     string  `json:"address,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`

	// Phone/Website/OpeningHours are whatever OSM/Google happen to have --
	// entirely optional and inconsistent between places. SuggestedCategory
	// is a best-effort guess (OSM's own amenity/shop tag, or Google's place
	// type) at one of our own Category keys, letting the client offer a
	// single-tap "poleć" instead of always forcing a category picker; ""
	// means no confident guess, client asks the parent instead.
	Phone             string `json:"phone,omitempty"`
	Website           string `json:"website,omitempty"`
	OpeningHours      string `json:"opening_hours,omitempty"`
	SuggestedCategory string `json:"suggested_category,omitempty"`
}

type SearchResponse struct {
	Candidates []SearchCandidate `json:"candidates,omitempty"`
	// Provider is which search actually served [Candidates] -- SourceGoogle or SourceOSM (see
	// places.SearchWithFallback) -- purely informational, e.g. so the client can show a subtle
	// "wyniki z OpenStreetMap" hint when the shared Google budget ran out for the day/month.
	Provider string `json:"provider,omitempty"`
}
