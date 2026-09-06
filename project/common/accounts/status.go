package accountspg

type Status int

const (
	StatusUnknown Status = iota
	StatusEnabled
	StatusDisabled
	StatusBanned
)

func (a Status) String() string {
	switch a {
	case StatusEnabled:
		return "Enabled"
	case StatusDisabled:
		return "Disabled"
	case StatusBanned:
		return "Banned"
	default:
		return "Unknown"
	}
}

func (a Status) Ordinal() int {
	return int(a)
}
