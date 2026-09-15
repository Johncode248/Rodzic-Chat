package placespg

import "errors"

// RecommendRequest either recommends an existing Place (PlaceId set), or
// describes a brand new one to create-then-recommend in the same call --
// either picked from a SearchCandidate (Source="osm", ExternalRef set) or
// typed by hand (Source="manual", ExternalRef empty). Coordinates are
// optional either way: a manually-added place with no known address can
// still be recommended, it just never gets a "Pokaż na mapie" button (see
// Place.HasCoordinates).
type RecommendRequest struct {
	PlaceId string `json:"place_id,omitempty"`

	Source       string   `json:"source,omitempty"`
	ExternalRef  string   `json:"external_ref,omitempty"`
	Name         string   `json:"name,omitempty"`
	Category     string   `json:"category,omitempty"`
	Address      string   `json:"address,omitempty"`
	Latitude     *float64 `json:"latitude,omitempty"`
	Longitude    *float64 `json:"longitude,omitempty"`
	Phone        string   `json:"phone,omitempty"`
	Website      string   `json:"website,omitempty"`
	OpeningHours string   `json:"opening_hours,omitempty"`

	Comment string `json:"comment,omitempty"`
}

var (
	ErrRecommendMissingTarget   = errors.New("either place_id or a new place's name+category must be given")
	ErrRecommendInvalidCategory = errors.New("unknown category")
)

func (r RecommendRequest) Validate() error {
	if r.PlaceId != "" {
		return nil
	}
	if r.Name == "" || r.Category == "" {
		return ErrRecommendMissingTarget
	}
	if !IsValidCategory(r.Category) {
		return ErrRecommendInvalidCategory
	}
	return nil
}
