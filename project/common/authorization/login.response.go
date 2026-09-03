package authorizationpg

type LoginResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
}

func (reply LoginReply) Inherit() LoginResponse {
	return LoginResponse{
		AccessToken:  reply.AccessToken,
		RefreshToken: reply.RefreshToken,
		ExpiresIn:    reply.ExpiresIn,
	}
}
