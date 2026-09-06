package eventspg

type RegistrationUpdateRequest struct {
	Status       *RegistrationStatus `json:"status,omitempty"`
	ConsentGiven *bool               `json:"consent_given,omitempty"`
}
