package authorization

import (
	"context"
	"errors"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/credentials"
)

var ErrProviderAccountNotFound = errors.New("no account exists for this email yet -- ask your placówka's director for an invite link first")

// LoginWithVerifiedEmail issues tokens for the Account matching argument.Email
// -- no password check, because the caller (login.google.mobile.go /
// login.apple.mobile.go) has already verified a signed provider token
// proving the caller controls that email. Never creates an account: a
// parent/teacher/director always has to exist already (via a director's
// registration link, or the original director /register), this only ever
// links a *login*, matching the exact same "find by email" rule the old
// web Google-login redirect flow already used.
func LoginWithVerifiedEmail(ctx context.Context, argument authorizationpg.LoginProviderArgument) (res authorizationpg.LoginReply, err error) {
	ctx = logger.Enter(ctx, "authorization.LoginWithVerifiedEmail", argument.Email)
	defer logger.Exit(ctx, "authorization.LoginWithVerifiedEmail", &res, &err)

	accountList, err := gormdb.Select[accountspg.Account](ctx, databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "email", Operator: "=", Value: argument.Email}}},
		},
	})
	if err != nil {
		return res, err
	}
	if len(accountList) == 0 {
		return res, ErrProviderAccountNotFound
	}

	issued, err := credentials.IssueTokens(ctx, credentials.IssueTokensArgument{
		AccountId:  accountList[0].Id,
		Email:      accountList[0].Email,
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
