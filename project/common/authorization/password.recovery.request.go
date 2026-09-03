package authorizationpg

type PasswordRecoveryRequest struct {
	Email string `json:"email,omitempty"`
}
