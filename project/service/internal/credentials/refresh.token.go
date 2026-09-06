package credentials

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	sessionspg "github.com/Johncode248/Rodzic-Chat/project/common/sessions"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/sessions"
)

func RefreshToken(ctx context.Context, argument authorizationpg.RefreshTokenArgument) (res authorizationpg.RefreshTokenReply, err error) {
	ctx = logger.Enter(ctx, "credentials.RefreshToken")
	defer logger.Exit(ctx, "credentials.RefreshToken", &err)

	token, err := ExtractToken(ctx, argument.RefreshToken)
	if err != nil {
		return authorizationpg.RefreshTokenReply{}, err
	}

	if token.TokenKind != "refresh" || token.SessionId == "" {
		return authorizationpg.RefreshTokenReply{}, errors.New("not a refresh token")
	}

	account, err := gormdb.Select[accountspg.Account](ctx, databasepg.Query{
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

	sessionResult, err := sessions.Select(ctx, sessionspg.SelectArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "id", Operator: "=", Value: token.SessionId},
					},
				},
			},
		},
	})
	if err != nil {
		return authorizationpg.RefreshTokenReply{}, err
	}

	if len(sessionResult.Sessions) == 0 {
		return authorizationpg.RefreshTokenReply{}, errors.New("session not found")
	}

	session := sessionResult.Sessions[0]

	if session.RevokedAt != nil {
		return authorizationpg.RefreshTokenReply{}, errors.New("session revoked")
	}

	if session.ExpiresAt.Before(time.Now()) {
		return authorizationpg.RefreshTokenReply{}, errors.New("session expired")
	}

	// Hash mismatch means the presented refresh token isn't the one this
	// session was last issued -- either a stale/rotated-away token or a
	// stolen one being replayed. Neither should be honored.
	if HashToken(argument.RefreshToken) != session.RefreshTokenHash {
		return authorizationpg.RefreshTokenReply{}, errors.New("refresh token mismatch")
	}

	// Carry the token's current active org forward across the refresh, so
	// refreshing silently doesn't reset a switched-to org back to primary.
	accessToken, err := GenerateAccessToken(ctx, token.UserId, token.Email, session.Id, token.OrganizationId)
	if err != nil {
		return authorizationpg.RefreshTokenReply{}, err
	}

	refreshToken, err := GenerateRefreshToken(ctx, token.UserId, token.Email, session.Id, token.OrganizationId)
	if err != nil {
		return authorizationpg.RefreshTokenReply{}, err
	}

	// Rotate in place: update the same session row rather than creating a
	// new one, since iOS refreshes silently many times per login and each
	// device should stay exactly one row.
	newHash := HashToken(refreshToken)
	newExpiresAt := time.Now().Add(RefreshTokenTTL)
	now := time.Now()

	if _, err = sessions.Update(ctx, sessionspg.UpdateArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "id", Operator: "=", Value: session.Id},
					},
				},
			},
		},
		RefreshTokenHash: &newHash,
		ExpiresAt:        &newExpiresAt,
		LastUsedAt:       &now,
	}); err != nil {
		return authorizationpg.RefreshTokenReply{}, err
	}

	return authorizationpg.RefreshTokenReply{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(AccessTokenTTL.Seconds()),
		ExpiresAt:    time.Now().Add(AccessTokenTTL).Unix(),
	}, nil
}
