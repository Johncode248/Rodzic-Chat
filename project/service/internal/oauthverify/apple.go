package oauthverify

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

var appleKeys = newKeyCache("https://appleid.apple.com/auth/keys")

// VerifyAppleIdentityToken verifies a Sign in with Apple identity token
// (returned by AuthenticationServices on iOS) and returns the verified
// email. audience must be the app's bundle identifier -- that's what
// native (non-web) Sign in with Apple puts in the token's "aud" claim.
func VerifyAppleIdentityToken(identityToken, audience string) (email string, err error) {
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(identityToken, claims, func(t *jwt.Token) (interface{}, error) {
		kid, _ := t.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("oauthverify: token has no kid header")
		}
		return appleKeys.publicKey(kid)
	})
	if err != nil {
		return "", fmt.Errorf("oauthverify: invalid apple identity_token: %w", err)
	}
	if !token.Valid {
		return "", errors.New("oauthverify: invalid apple identity_token")
	}

	iss, _ := claims["iss"].(string)
	if iss != "https://appleid.apple.com" {
		return "", fmt.Errorf("oauthverify: unexpected apple issuer %q", iss)
	}

	aud, _ := claims["aud"].(string)
	if aud != audience {
		return "", errors.New("oauthverify: apple identity_token audience mismatch")
	}

	if !parseVerifiedFlag(claims["email_verified"]) {
		return "", ErrEmailNotVerified
	}

	email, _ = claims["email"].(string)
	if email == "" {
		return "", errors.New("oauthverify: apple identity_token has no email claim")
	}

	return email, nil
}

// parseVerifiedFlag handles both shapes an "email_verified" claim shows up
// as in the wild: a real JSON bool (Google's documented shape, and Apple's
// in some SDK versions), or the string "true"/"false" (Apple's shape in
// other SDK versions). Shared by both providers' verifiers -- see
// google.go.
func parseVerifiedFlag(raw interface{}) bool {
	switch v := raw.(type) {
	case bool:
		return v
	case string:
		return v == "true"
	default:
		return false
	}
}
