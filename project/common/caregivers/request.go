package caregiverspg

import "time"

// UpsertRequest creates or updates the caller's own Caregiver profile for
// a given Type -- there's at most one per (account, type), see Caregiver's
// unique index.
type UpsertRequest struct {
	Type           int    `json:"type,omitempty"`
	DisplayName    string `json:"display_name,omitempty"`
	Description    string `json:"description,omitempty"`
	Experience     string `json:"experience,omitempty"`
	HourlyRate     int64  `json:"hourly_rate,omitempty"`
	City           string `json:"city,omitempty"`
	Phone          string `json:"phone,omitempty"`
	Email          string `json:"email,omitempty"`
	Subjects       string `json:"subjects,omitempty"`
	Online         bool   `json:"online"`
	Stationary     bool   `json:"stationary"`
	Travel         bool   `json:"travel"`
	TravelRadiusKm int    `json:"travel_radius_km,omitempty"`
	ServiceCities  string `json:"service_cities,omitempty"`
}

// Summary is one row in the browse list -- the full Caregiver plus its
// recommendation tally, so the list doesn't need a second round trip per
// row just to show "X poleca".
type Summary struct {
	Caregiver           Caregiver `json:"caregiver,omitempty"`
	RecommendationCount int64     `json:"recommendation_count,omitempty"`
}

type ListResponse struct {
	Caregivers []Summary `json:"caregivers,omitempty"`
}

// Detail is the single-record aggregate: the Caregiver, its tally, and
// two viewer-relative flags the client needs to decide what to show --
// "Polecam" only makes sense if this isn't the caller's own profile
// (IsMine), and RecommendedByMe drives whether that button reads
// "Polecam" or "Polecono" (already endorsed).
// Recommender is one "Polecam" on a profile: who and what they said.
type Recommender struct {
	Name       string    `json:"name,omitempty"`
	Comment    string    `json:"comment,omitempty"`
	CreateTime time.Time `json:"create_time,omitempty"`
}

type Detail struct {
	Caregiver           Caregiver     `json:"caregiver,omitempty"`
	RecommendationCount int64         `json:"recommendation_count,omitempty"`
	RecommendedByMe     bool          `json:"recommended_by_me,omitempty"`
	IsMine              bool          `json:"is_mine,omitempty"`
	Recommenders        []Recommender `json:"recommenders,omitempty"`
}

type RecommendRequest struct {
	CaregiverId string `json:"caregiver_id,omitempty"`
	Comment     string `json:"comment,omitempty"`
}
