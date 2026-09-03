package credentials

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	userspg "github.com/Johncode248/Rodzic-Chat/project/common/users"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func GenerateAccessToken(ctx context.Context, userId, email string) (res string, err error) {
	ctx = logger.Enter(ctx, "credentials.GenerateAccessToken")
	defer logger.Exit(ctx, "credentials.GenerateAccessToken", &err)

	userTemp, err := gormdb.Select[userspg.User](ctx, databasepg.Query{
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

	if len(userTemp) == 0 {
		return res, errors.New("account not found")
	}

	claims2 := MyCustomClaims{
		UserId:    userId,
		SessionId: uuid.NewString(), // trzeba zamienic na prawdziwe session id work wazne
		Email:     email,
		TokenKind: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
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
