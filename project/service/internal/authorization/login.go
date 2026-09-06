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

	if len(accountList) == 0 {
		return res, errors.New("account not found")
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
		return res, errors.New("account not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(credentialList[0].Password), []byte(argument.Password))
	if err != nil {
		return authorizationpg.LoginReply{}, errors.New("password incorrect")
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
