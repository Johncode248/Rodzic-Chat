package credentials

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	userspg "github.com/Johncode248/Rodzic-Chat/project/common/users"
)

func GenerateRefreshToken(ctx context.Context, userId, email string) (res string, err error) {
	ctx = logger.Enter(ctx, "credentials.GenerateRefreshToken")
	defer logger.Exit(ctx, "credentials.GenerateRefreshToken", &err)

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
		Email:     email,
		TokenKind: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(720 * time.Hour)),
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
