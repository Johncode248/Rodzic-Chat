package authorization

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	credentialspg "github.com/Johncode248/Rodzic-Chat/project/common/credentials"
	userspg "github.com/Johncode248/Rodzic-Chat/project/common/users"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/credentials"
	"golang.org/x/crypto/bcrypt"
)

func Login(ctx context.Context, argument authorizationpg.LoginArgument) (res authorizationpg.LoginReply, err error) {
	ctx = logger.Enter(ctx, "authorization.Login", argument)
	defer logger.Exit(ctx, "authorization.Login", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	users, err := gormdb.Select[userspg.User](ctx, databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{
				Filters: []databasepg.Filter{
					{
						Key:      "email",
						Operator: "=",
						Value:    argument.Email,
					},
				},
			},
		},
	})
	if err != nil {
		return res, err
	}

	if len(users) == 0 {
		return res, errors.New("user not found")
	}

	credentialList, err := gormdb.Select[credentialspg.Credential](ctx, databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{
				Filters: []databasepg.Filter{
					{
						Key:      "user_id",
						Operator: "=",
						Value:    users[0].Id,
					},
				},
			},
		},
	})
	if err != nil {
		return res, err
	}

	if len(credentialList) == 0 {
		return res, errors.New("user not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(credentialList[0].Password), []byte(argument.Password))
	if err != nil {
		return authorizationpg.LoginReply{}, errors.New("password incorrect")
	}

	accessToken, err := credentials.GenerateAccessToken(ctx, users[0].Id, users[0].Email)
	if err != nil {
		return res, err
	}

	refreshToken, err := credentials.GenerateRefreshToken(ctx, users[0].Id, users[0].Email)
	if err != nil {
		return res, err
	}

	return authorizationpg.LoginReply{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Hour * 23).Unix(),
	}, nil
}
