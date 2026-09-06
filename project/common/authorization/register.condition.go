package authorizationpg

import (
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	credentialspg "github.com/Johncode248/Rodzic-Chat/project/common/credentials"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
)

type RegisterCondition struct {
	Account      accountspg.Account
	Organization organizationspg.Organization
	Credential   credentialspg.Credential
	AccessToken  string
	RefreshToken string
}
