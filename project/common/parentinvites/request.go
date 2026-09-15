package parentinvitespg

type CreateResponse struct {
	Token     string `json:"token,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

// PublicInfoResponse never errors on an invalid/expired/used token --
// Usable=false + Reason renders more easily on a client than a raw HTTP
// error, same reasoning as registrationlinks.PublicInfoResponse.
type PublicInfoResponse struct {
	InviterName string `json:"inviter_name,omitempty"`
	Usable      bool   `json:"usable"`
	Reason      string `json:"reason,omitempty"`
}

// RegisterRequest is for someone with NO account yet accepting an invite --
// creates the account then links it. Someone who already has an account
// instead calls the authenticated accept endpoint with no body.
type RegisterRequest struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

// AcceptResponse reports which children actually got linked -- a child
// already at its two-parent cap (see project/common/registrationlinks'
// same rule) is silently skipped rather than failing the whole accept, so
// SkippedChildIds lets the client explain why the count is lower than
// expected.
type AcceptResponse struct {
	LinkedChildIds  []string `json:"linked_child_ids,omitempty"`
	SkippedChildIds []string `json:"skipped_child_ids,omitempty"`
}
