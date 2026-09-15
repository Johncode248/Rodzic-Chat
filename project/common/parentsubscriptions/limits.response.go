package parentsubscriptionspg

// LimitsResponse is the consolidated "Limity" screen's data -- everything a parent's usage is
// capped by across the app, in one call, rather than three separate ones (AI, places, Google
// search). *Limit fields of -1 mean unlimited (only ever PlanMax, unreachable in MVP since only
// PlanFree is selectable -- see PlanView on the client).
type LimitsResponse struct {
	Plan string `json:"plan,omitempty"`

	AiTokensUsed  int64  `json:"ai_tokens_used"`
	AiTokensLimit int64  `json:"ai_tokens_limit"`
	PeriodResetAt string `json:"period_reset_at,omitempty"`

	RecommendationsUsed  int64 `json:"recommendations_used"`
	RecommendationsLimit int64 `json:"recommendations_limit"`

	GooglePlacesSearchesToday int64 `json:"google_places_searches_today"`
	GooglePlacesDailyLimit    int64 `json:"google_places_daily_limit"`
}
