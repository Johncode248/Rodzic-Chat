package caregiverspg

// CaregiverType must always be appended after the last existing value,
// never inserted -- this is stored as a plain int ordinal in the DB (see
// project/common/chats/channel.type.go for the same convention and why).
type CaregiverType int

const (
	CaregiverTypeUnknown CaregiverType = iota
	CaregiverTypeNanny
	CaregiverTypeTutor
)

func (c CaregiverType) String() string {
	switch c {
	case CaregiverTypeNanny:
		return "Nanny"
	case CaregiverTypeTutor:
		return "Tutor"
	default:
		return "Unknown"
	}
}

func (c CaregiverType) Ordinal() int {
	return int(c)
}

func CaregiverTypeFromOrdinal(v int) CaregiverType {
	return CaregiverType(v)
}
