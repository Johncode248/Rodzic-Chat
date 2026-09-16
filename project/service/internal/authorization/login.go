package authorization

import (
	"context"
	"errors"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	credentialspg "github.com/Johncode248/Rodzic-Chat/project/common/credentials"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/credentials"
	"golang.org/x/crypto/bcrypt"
)

// errInvalidCredentials is returned for both "no such account" and "wrong
// password" -- a distinct message per case would let a caller enumerate
// which emails are registered by watching which error comes back.
var errInvalidCredentials = errors.New("invalid credentials")

// dummyPasswordHash is a real bcrypt hash of an arbitrary fixed string,
// used only to burn roughly the same CPU time as a genuine
// bcrypt.CompareHashAndPassword call (see its two call sites below) when
// there is no real Credential to compare against. Without this, "no such
// account"/"no credential row" return in a handful of DB-query
// milliseconds while a real "wrong password" takes tens of milliseconds
// longer for bcrypt itself to run -- an identical error MESSAGE (see
// errInvalidCredentials) doesn't hide that LATENCY difference, which is
// enough to enumerate registered logins by timing alone.
const dummyPasswordHash = "$2a$10$CwTycUXWue0Thq9StjUM0uJ8Q0HmMx4gRe4jtdNQVWTMkuZ8Yn6uS"

func Login(ctx context.Context, argument authorizationpg.LoginArgument) (res authorizationpg.LoginReply, err error) {
	ctx = logger.Enter(ctx, "authorization.Login", argument)
	defer logger.Exit(ctx, "authorization.Login", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	accountList, err := gormdb.Select[accountspg.Account](ctx, databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{
				Filters: []databasepg.Filter{
					{
						Key:      "login",
						Operator: "=",
						Value:    argument.Login,
					},
				},
			},
		},
	})
	if err != nil {
		return res, err
	}

	// Deliberately the same error for "no such account" and "wrong
	// password" below (see errInvalidCredentials) -- a distinct message
	// per case lets an attacker enumerate which emails are registered.
	if len(accountList) == 0 {
		_ = bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte(argument.Password))
		return res, errInvalidCredentials
	}

	credentialList, err := gormdb.Select[credentialspg.Credential](ctx, databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{
				Filters: []databasepg.Filter{
					{
						Key:      "user_id",
						Operator: "=",
						Value:    accountList[0].Id,
					},
				},
			},
		},
	})
	if err != nil {
		return res, err
	}

	if len(credentialList) == 0 {
		_ = bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte(argument.Password))
		return res, errInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(credentialList[0].Password), []byte(argument.Password))
	if err != nil {
		return authorizationpg.LoginReply{}, errInvalidCredentials
	}

	issued, err := credentials.IssueTokens(ctx, credentials.IssueTokensArgument{
		AccountId:  accountList[0].Id,
		Email:      accountList[0].Login,
		DeviceId:   argument.DeviceId,
		DeviceName: argument.DeviceName,
		UserAgent:  argument.UserAgent,
		IpAddress:  argument.IpAddress,
	})
	if err != nil {
		return res, err
	}

	return authorizationpg.LoginReply{
		AccessToken:  issued.AccessToken,
		RefreshToken: issued.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(credentials.AccessTokenTTL.Seconds()),
		ExpiresAt:    issued.ExpiresAt.Unix(),
	}, nil
}
