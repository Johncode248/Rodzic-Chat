package passwordresetspg

type SelectResponse struct {
	PasswordResets []PasswordReset `json:"password_resets,omitempty"`
}
