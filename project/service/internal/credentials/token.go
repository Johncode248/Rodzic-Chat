package credentials

import "github.com/Johncode248/Rodzic-Chat/variables"

var sampleSecretKey = []byte(variables.JWTSecret)

type Token struct {
	UserId         string `json:"userId"`
	SessionId      string `json:"sessionId"`
	Email          string `json:"email"`
	OrganizationId string `json:"organizationId"`
	AccountType    int    `json:"accountType"`
	TokenKind      string `json:"tokenKind"`
}
