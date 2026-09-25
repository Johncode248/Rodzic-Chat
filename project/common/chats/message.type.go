package chatspg

type MessageType int

const (
	MessageTypeUnknown MessageType = iota
	MessageTypeText
	MessageTypePoll
	MessageTypeSystem
	// Appended after the last value (stored as an int ordinal -- never insert in the middle).
	MessageTypeConsent
	MessageTypeSignup
)

func (m MessageType) String() string {
	switch m {
	case MessageTypeText:
		return "Text"
	case MessageTypePoll:
		return "Poll"
	case MessageTypeSystem:
		return "System"
	case MessageTypeConsent:
		return "Consent"
	case MessageTypeSignup:
		return "Signup"
	default:
		return "Unknown"
	}
}

func (m MessageType) Ordinal() int {
	return int(m)
}
