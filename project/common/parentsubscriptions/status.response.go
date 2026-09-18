package parentsubscriptionspg

// StatusResponse is what a parent's app reads to render their plan. No usage/limit fields for
// MVP -- see ParentSubscription's doc comment.
type StatusResponse struct {
	Plan string `json:"plan,omitempty"`
}

// TestSetPlanRequest: see transport/parentsubscriptions/test.set.plan.go
// for why this exists and how it's gated.
type TestSetPlanRequest struct {
	Plan string `json:"plan,omitempty"`
}
