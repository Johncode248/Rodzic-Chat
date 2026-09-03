package credentials

import (
	"context"
	"errors"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	userspg "github.com/Johncode248/Rodzic-Chat/project/common/users"
)

func RefreshToken(ctx context.Context, argument authorizationpg.RefreshTokenArgument) (res authorizationpg.RefreshTokenReply, err error) {
	ctx = logger.Enter(ctx, "credentials.RefreshToken")
	defer logger.Exit(ctx, "credentials.RefreshToken", &err)

	token, err := ExtractToken(ctx, argument.RefreshToken)
	if err != nil {
		return authorizationpg.RefreshTokenReply{}, err
	}

	account, err := gormdb.Select[userspg.User](ctx, databasepg.Query{
		Groups: []databasepg.Group{
			{
				Filters: []databasepg.Filter{
					{
						Key:      "id",
						Operator: "=",
						Value:    token.UserId,
					},
				},
			},
		},
	})
	if err != nil {
		return authorizationpg.RefreshTokenReply{}, err
	}

	if len(account) == 0 {
		return authorizationpg.RefreshTokenReply{}, errors.New("account not found")
	}

	accessToken, err := GenerateAccessToken(ctx, token.UserId, token.Email)
	if err != nil {
		return authorizationpg.RefreshTokenReply{}, err
	}

	refreshToken, err := GenerateRefreshToken(ctx, token.UserId, token.Email)
	if err != nil {
		return authorizationpg.RefreshTokenReply{}, err
	}

	return authorizationpg.RefreshTokenReply{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "",
		ExpiresIn:    0,
		ExpiresAt:    0,
	}, nil
}
