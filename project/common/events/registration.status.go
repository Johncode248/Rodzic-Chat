package eventspg

type RegistrationStatus int

const (
	RegistrationStatusUnknown RegistrationStatus = iota
	RegistrationStatusConfirmed
	RegistrationStatusDeclined
	RegistrationStatusPending
)

func (r RegistrationStatus) String() string {
	switch r {
	case RegistrationStatusConfirmed:
		return "Confirmed"
	case RegistrationStatusDeclined:
		return "Declined"
	case RegistrationStatusPending:
		return "Pending"
	default:
		return "Unknown"
	}
}

func (r RegistrationStatus) Ordinal() int {
	return int(r)
}
