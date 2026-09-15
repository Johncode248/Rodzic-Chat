package parentsubscriptionspg

// StatusResponse is what a parent's app reads to render their plan +
// usage. AiRequestsLimit/AiTokensLimit are -1 for unlimited (PlanMax).
// AiTokensUsed/AiTokensLimit are the REAL enforcement metric (see
// ParentSubscription.AiTokensUsed's doc comment) -- AiRequestsUsed/Limit are kept only as a
// secondary, less precise stat; a client must gate "can the parent still send a message" on
// tokens, not requests.
type StatusResponse struct {
	Plan            string `json:"plan,omitempty"`
	AiRequestsUsed  int64  `json:"ai_requests_used"`
	AiRequestsLimit int64  `json:"ai_requests_limit"`
	AiTokensUsed    int64  `json:"ai_tokens_used"`
	AiTokensLimit   int64  `json:"ai_tokens_limit"`
	PeriodResetAt   string `json:"period_reset_at,omitempty"`
}

// TestSetPlanRequest: see transport/parentsubscriptions/test.set.plan.go
// for why this exists and how it's gated.
type TestSetPlanRequest struct {
	Plan string `json:"plan,omitempty"`
}
