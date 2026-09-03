package credentialspg

type SelectResponse struct {
	Credentials []Credential `json:"credentials,omitempty"`
}
