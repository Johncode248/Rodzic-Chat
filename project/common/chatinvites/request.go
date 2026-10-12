package chatinvitespg

import chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"

type CreateResponse struct {
	// Pin is the plaintext 8-digit PIN -- returned exactly once, here.
	Pin       string          `json:"pin,omitempty"`
	ExpiresAt string          `json:"expires_at,omitempty"`
	Channel   chatspg.Channel `json:"channel,omitempty"`
}

// LookupRequest is POST /private/chat-invites/lookup's body --
// authenticated, rate-limited, does not consume the invite.
type LookupRequest struct {
	Pin string `json:"pin,omitempty"`
}

// LookupResponse never errors on an invalid/expired/used PIN --
// Usable=false + Reason renders more easily on a client than a raw HTTP
// error.
type LookupResponse struct {
	InviterName string `json:"inviter_name,omitempty"`
	Usable      bool   `json:"usable"`
	Reason      string `json:"reason,omitempty"`
}

// AcceptRequest is POST /private/chat-invites/accept's body --
// authenticated, rate-limited.
type AcceptRequest struct {
	Pin string `json:"pin,omitempty"`
}

type AcceptResponse struct {
	Channel chatspg.Channel `json:"channel,omitempty"`
}
