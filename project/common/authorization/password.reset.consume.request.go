package authorizationpg

type ConsumePasswordResetRequest struct {
	Token       string `json:"token,omitempty"`
	NewPassword string `json:"new_password,omitempty"`
}
