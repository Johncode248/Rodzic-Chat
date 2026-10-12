package chatspg

import "time"

// ChannelMemberInfo is one row of a channel's member roster -- the raw
// ChannelMember plus the account's display name, so the client doesn't
// need a second round trip per member just to show who's who.
type ChannelMemberInfo struct {
	AccountId   string            `json:"account_id,omitempty"`
	DisplayName string            `json:"display_name,omitempty"`
	Role        ChannelMemberRole `json:"role,omitempty"`
	// Permissions are the member's EFFECTIVE permissions (role preset + individual grants);
	// GrantedPermissions only the individual bits (what the admin toggled).
	Permissions        ChannelPermission `json:"permissions,omitempty"`
	GrantedPermissions ChannelPermission `json:"granted_permissions,omitempty"`
	// BannedUntil is set only while a ban is active, and only for callers allowed to see bans.
	// Banned without an end date is signalled by BannedForever.
	BannedUntil   *time.Time `json:"banned_until,omitempty"`
	BannedForever bool       `json:"banned_forever,omitempty"`
}

type ChannelMemberListResponse struct {
	Members []ChannelMemberInfo `json:"members,omitempty"`
	// The caller's own role and effective permissions, so the client knows which actions to offer.
	MyRole        ChannelMemberRole `json:"my_role,omitempty"`
	MyPermissions ChannelPermission `json:"my_permissions,omitempty"`
}

// SetMemberRoleRequest changes another member's role and/or granted permissions (admin only).
// Role must be member or moderator; Permissions are the individually granted bits.
type SetMemberRoleRequest struct {
	ChannelId   string             `json:"channel_id,omitempty"`
	AccountId   string             `json:"account_id,omitempty"`
	Role        *ChannelMemberRole `json:"role,omitempty"`
	Permissions *ChannelPermission `json:"permissions,omitempty"`
}

// BanRequest blocks a member; DurationMinutes 0 = until lifted (max 366 days otherwise).
type BanRequest struct {
	ChannelId       string `json:"channel_id,omitempty"`
	AccountId       string `json:"account_id,omitempty"`
	DurationMinutes int    `json:"duration_minutes,omitempty"`
	Reason          string `json:"reason,omitempty"`
}
