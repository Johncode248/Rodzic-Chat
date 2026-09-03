package credentials

import (
	"context"
	"errors"
	"strings"

	"github.com/Johncode248/Rodzic-Chat/logger"
)

func ExtractToken(ctx context.Context, tokenString string) (t Token, err error) {
	ctx = logger.Enter(ctx, "credentials.ExtractToken")
	defer logger.Exit(ctx, "credentials.ExtractToken", &err)

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return sampleSecretKey, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		return Token{}, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return Token{}, errors.New("failed to read token claims")
	}

	var result Token
	if claims["tokenKind"].(string) != "access" {
		result = Token{
			UserId:    claims["userId"].(string),
			SessionId: claims["sessionId"].(string),
			Email:     claims["email"].(string),
			//CustomerId: claims["customerId"].(string),
		}
	} else {
		result = Token{
			UserId: claims["userId"].(string),
			Email:  claims["email"].(string),
			//CustomerId: claims["customerId"].(string),
		}
	}

	return result, nil
}
