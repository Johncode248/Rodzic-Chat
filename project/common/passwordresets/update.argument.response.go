package passwordresetspg

type UpdateResponse struct {
	PasswordReset PasswordReset `json:"password_reset,omitempty"`
}
