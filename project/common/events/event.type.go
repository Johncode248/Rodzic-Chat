package eventspg

type EventType int

const (
	EventTypeUnknown EventType = iota
	EventTypeGeneral
	EventTypeTrip
)

func (e EventType) String() string {
	switch e {
	case EventTypeGeneral:
		return "General"
	case EventTypeTrip:
		return "Trip"
	default:
		return "Unknown"
	}
}

func (e EventType) Ordinal() int {
	return int(e)
}
