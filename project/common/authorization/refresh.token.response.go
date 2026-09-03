package authorizationpg

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
}

func (reply RefreshTokenReply) Inherit() RefreshTokenResponse {
	return RefreshTokenResponse{
		AccessToken:  reply.AccessToken,
		RefreshToken: reply.RefreshToken,
		ExpiresIn:    reply.ExpiresIn,
	}
}
