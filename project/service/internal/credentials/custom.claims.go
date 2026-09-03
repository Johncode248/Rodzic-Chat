package credentials

import "github.com/golang-jwt/jwt/v4"

type MyCustomClaims struct {
	UserId    string `json:"userId"`
	SessionId string `json:"sessionId"`
	Email     string `json:"email"`
	TokenKind string `json:"tokenKind"`
	jwt.RegisteredClaims
}
