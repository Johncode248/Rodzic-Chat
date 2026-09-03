package authorizationpg

import (
	credentialspg "github.com/Johncode248/Rodzic-Chat/project/common/credentials"
	userspg "github.com/Johncode248/Rodzic-Chat/project/common/users"
)

type RegisterCondition struct {
	User         userspg.User
	Credential   credentialspg.Credential
	AccessToken  string
	RefreshToken string
}
