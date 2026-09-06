package chatspg

type MessageType int

const (
	MessageTypeUnknown MessageType = iota
	MessageTypeText
	MessageTypePoll
	MessageTypeSystem
)

func (m MessageType) String() string {
	switch m {
	case MessageTypeText:
		return "Text"
	case MessageTypePoll:
		return "Poll"
	case MessageTypeSystem:
		return "System"
	default:
		return "Unknown"
	}
}

func (m MessageType) Ordinal() int {
	return int(m)
}
