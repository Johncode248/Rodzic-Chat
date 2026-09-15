package chatspg

// ChannelMemberRole must always be appended after the last existing value,
// never inserted -- stored as a plain int ordinal, see channel.type.go for
// the same convention.
type ChannelMemberRole int

const (
	// ChannelMemberRoleMember can post and leave, nothing else.
	ChannelMemberRoleMember ChannelMemberRole = iota
	// ChannelMemberRoleAdmin is granted to whoever created the channel via
	// a chat invite (see project/common/chatinvites) -- can invite more
	// people and remove other members, in addition to everything a plain
	// member can do.
	ChannelMemberRoleAdmin
)

func (r ChannelMemberRole) Ordinal() int {
	return int(r)
}

func ChannelMemberRoleFromOrdinal(v int) ChannelMemberRole {
	return ChannelMemberRole(v)
}
