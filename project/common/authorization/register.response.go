package authorizationpg

type RegisterResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
}

func (reply RegisterReply) Inherit() RegisterResponse {
	return RegisterResponse{
		AccessToken:  reply.AccessToken,
		RefreshToken: reply.RefreshToken,
		ExpiresIn:    reply.ExpiresIn,
	}
}
