package credentials

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func signClaims(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(sampleSecretKey)
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}
	return signed
}

func TestExtractToken_ValidAccessToken(t *testing.T) {
	signed := signClaims(t, jwt.MapClaims{
		"userId":         "user-1",
		"sessionId":      "session-1",
		"email":          "parent@example.com",
		"organizationId": "org-1",
		"accountType":    float64(1),
		"tokenKind":      "access",
		"exp":            time.Now().Add(time.Hour).Unix(),
	})

	result, err := ExtractToken(context.Background(), signed)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.UserId != "user-1" || result.SessionId != "session-1" || result.Email != "parent@example.com" ||
		result.OrganizationId != "org-1" || result.AccountType != 1 || result.TokenKind != "access" {
		t.Fatalf("unexpected extracted token: %+v", result)
	}
}

func TestExtractToken_MissingUserId_ReturnsErrorNotPanic(t *testing.T) {
	signed := signClaims(t, jwt.MapClaims{
		"tokenKind": "access",
		"exp":       time.Now().Add(time.Hour).Unix(),
	})

	if _, err := ExtractToken(context.Background(), signed); err == nil {
		t.Fatal("expected an error for a token missing userId, got nil")
	}
}

func TestExtractToken_WrongTypeClaims_ReturnsErrorNotPanic(t *testing.T) {
	// userId present but wrong type -- must not panic the request goroutine.
	signed := signClaims(t, jwt.MapClaims{
		"userId":    12345,
		"tokenKind": "access",
		"exp":       time.Now().Add(time.Hour).Unix(),
	})

	if _, err := ExtractToken(context.Background(), signed); err == nil {
		t.Fatal("expected an error for a userId claim with the wrong type, got nil")
	}
}

func TestExtractToken_ExpiredToken_ReturnsError(t *testing.T) {
	signed := signClaims(t, jwt.MapClaims{
		"userId": "user-1",
		"exp":    time.Now().Add(-time.Hour).Unix(),
	})

	if _, err := ExtractToken(context.Background(), signed); err == nil {
		t.Fatal("expected an error for an expired token, got nil")
	}
}
