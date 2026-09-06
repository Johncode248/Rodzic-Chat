package passwordresetspg

type CreateResponse struct {
	PasswordReset PasswordReset `json:"password_reset,omitempty"`
}
