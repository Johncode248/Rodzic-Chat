package credentials

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	"github.com/golang-jwt/jwt/v4"
)

// GenerateRefreshToken mints a refresh token tied to sessionId -- see
// GenerateAccessToken's doc comment (including requestedOrganizationId).
// The refresh token's SessionId claim is what lets RefreshToken
// (refresh.token.go) look up and rotate the correct sessionspg.Session row
// instead of operating statelessly.
func GenerateRefreshToken(ctx context.Context, userId, email, sessionId, requestedOrganizationId string) (res string, err error) {
	ctx = logger.Enter(ctx, "credentials.GenerateRefreshToken")
	defer logger.Exit(ctx, "credentials.GenerateRefreshToken", &err)

	accountTemp, err := gormdb.Select[accountspg.Account](ctx, databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{
				Filters: []databasepg.Filter{
					{
						Key:      "id",
						Operator: "=",
						Value:    userId,
					},
				},
			},
		},
	})
	if err != nil {
		return res, err
	}

	if len(accountTemp) == 0 {
		return res, errors.New("account not found")
	}

	activeOrganizationId, allOrganizationIds, err := resolveOrganizations(ctx, userId, accountTemp[0].OrganizationId, requestedOrganizationId)
	if err != nil {
		return res, err
	}

	claims2 := MyCustomClaims{
		UserId:          userId,
		SessionId:       sessionId,
		Email:           email,
		OrganizationId:  activeOrganizationId,
		OrganizationIds: allOrganizationIds,
		AccountType:     int(accountTemp[0].AccountType),
		TokenKind:       "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "test",
			Subject:   "somebody",
			ID:        "1",
			Audience:  []string{"somebody_else"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims2)

	tokenString, err := token.SignedString(sampleSecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
