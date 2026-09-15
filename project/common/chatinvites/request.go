package chatinvitespg

import chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"

type CreateResponse struct {
	Token     string          `json:"token,omitempty"`
	ExpiresAt string          `json:"expires_at,omitempty"`
	Channel   chatspg.Channel `json:"channel,omitempty"`
}

// PublicInfoResponse never errors on an invalid/expired/used token --
// Usable=false + Reason renders more easily on a client than a raw HTTP
// error, same reasoning as parentinvitespg.PublicInfoResponse.
type PublicInfoResponse struct {
	InviterName string `json:"inviter_name,omitempty"`
	Usable      bool   `json:"usable"`
	Reason      string `json:"reason,omitempty"`
}

type AcceptResponse struct {
	Channel chatspg.Channel `json:"channel,omitempty"`
}
