package eventspg

type RegistrationSelectResponse struct {
	EventRegistrations []EventRegistration `json:"event_registrations,omitempty"`
}
