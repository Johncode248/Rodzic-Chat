package parentinvitespg

type CreateResponse struct {
	// Pin is the plaintext 8-digit PIN -- returned exactly once, here.
	Pin       string `json:"pin,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

// AcceptRequest is POST /private/parent-invites/accept's body --
// authenticated, rate-limited (see module.go).
type AcceptRequest struct {
	Pin string `json:"pin,omitempty"`
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
