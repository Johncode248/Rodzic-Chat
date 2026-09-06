package eventspg

type RegistrationCreateRequest struct {
	EventId      string             `json:"event_id,omitempty"`
	ChildId      string             `json:"child_id,omitempty"`
	Status       RegistrationStatus `json:"status,omitempty"`
	ConsentGiven bool               `json:"consent_given,omitempty"`
}
