package chatspg

// ChannelMemberInfo is one row of a channel's member roster -- the raw
// ChannelMember plus the account's display name, so the client doesn't
// need a second round trip per member just to show who's who.
type ChannelMemberInfo struct {
	AccountId   string            `json:"account_id,omitempty"`
	DisplayName string            `json:"display_name,omitempty"`
	Role        ChannelMemberRole `json:"role,omitempty"`
}

type ChannelMemberListResponse struct {
	Members []ChannelMemberInfo `json:"members,omitempty"`
}
