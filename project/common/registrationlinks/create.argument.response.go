package registrationlinkspg

type CreateResponse struct {
	RegistrationLink RegistrationLink `json:"registration_link,omitempty"`
	// Pin is the plaintext 8-digit PIN -- returned exactly once, here, and
	// never retrievable again (only its hash is stored). The client is
	// responsible for any display formatting (e.g. "1234-5678").
	Pin string `json:"pin,omitempty"`
}
