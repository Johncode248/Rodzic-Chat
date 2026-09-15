package placespg

// PlaceWithStats is a Place plus the two things every list/detail view
// needs and that aren't stored on Place itself: how many parents recommend
// it, and whether the calling parent already has.
type PlaceWithStats struct {
	Place
	RecommendationCount int64 `json:"recommendation_count"`
	RecommendedByMe     bool  `json:"recommended_by_me"`
}

type ListResponse struct {
	Places []PlaceWithStats `json:"places,omitempty"`
}
