package chatspg

type ChannelType int

// ChannelTypeParents must always be appended after the last existing
// value, never inserted -- this is stored as a plain int ordinal in the
// DB, so inserting a new value in the middle would silently reinterpret
// every existing row's channel type.
const (
	ChannelTypeUnknown ChannelType = iota
	ChannelTypeAnnouncement
	ChannelTypeGroup
	ChannelTypeDirect
	ChannelTypeParents
)

func (c ChannelType) String() string {
	switch c {
	case ChannelTypeAnnouncement:
		return "Announcement"
	case ChannelTypeGroup:
		return "Group"
	case ChannelTypeDirect:
		return "Direct"
	case ChannelTypeParents:
		return "Parents"
	default:
		return "Unknown"
	}
}

func (c ChannelType) Ordinal() int {
	return int(c)
}
