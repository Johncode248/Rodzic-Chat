package authorizationpg

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}
