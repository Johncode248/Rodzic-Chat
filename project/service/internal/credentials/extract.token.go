package credentials

import (
	"context"
	"errors"
	"strings"

	"github.com/Johncode248/Rodzic-Chat/logger"
	"github.com/golang-jwt/jwt/v4"
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

	userId, ok := claims["userId"].(string)
	if !ok || userId == "" {
		return Token{}, errors.New("token missing userId claim")
	}

	email, _ := claims["email"].(string)
	sessionId, _ := claims["sessionId"].(string)
	tokenKind, _ := claims["tokenKind"].(string)

	result := Token{
		UserId:    userId,
		SessionId: sessionId,
		Email:     email,
		TokenKind: tokenKind,
	}

	if v, ok := claims["organizationId"].(string); ok {
		result.OrganizationId = v
	}

	if v, ok := claims["accountType"].(float64); ok {
		result.AccountType = int(v)
	}

	return result, nil
}
