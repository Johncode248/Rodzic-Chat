package eventspg

type RegistrationUpdateRequest struct {
	Status *RegistrationStatus `json:"status,omitempty"`
}
