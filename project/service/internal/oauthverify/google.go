package oauthverify

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

var googleKeys = newKeyCache("https://www.googleapis.com/oauth2/v3/certs")

// ErrEmailNotVerified means the provider itself hasn't confirmed the user
// controls this email address -- refusing to trust it for account lookup.
var ErrEmailNotVerified = errors.New("oauthverify: provider did not report the email as verified")

// VerifyGoogleIDToken verifies a Google Sign-In ID token's signature,
// issuer, audience and expiry (all standard JWT checks, done by
// jwt.ParseWithClaims against Google's published keys), then returns the
// verified email. audience must be the OAuth client ID the token was
// requested with -- for Android's native Google Sign-In this is
// conventionally the app's *web* client ID (Android passes it as
// "server client ID" precisely so the backend can verify against one
// stable audience regardless of package name/signing cert).
func VerifyGoogleIDToken(idToken, audience string) (email string, err error) {
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(idToken, claims, func(t *jwt.Token) (interface{}, error) {
		kid, _ := t.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("oauthverify: token has no kid header")
		}
		return googleKeys.publicKey(kid)
	})
	if err != nil {
		return "", fmt.Errorf("oauthverify: invalid google id_token: %w", err)
	}
	if !token.Valid {
		return "", errors.New("oauthverify: invalid google id_token")
	}

	iss, _ := claims["iss"].(string)
	if iss != "https://accounts.google.com" && iss != "accounts.google.com" {
		return "", fmt.Errorf("oauthverify: unexpected google issuer %q", iss)
	}

	aud, _ := claims["aud"].(string)
	if aud != audience {
		return "", errors.New("oauthverify: google id_token audience mismatch")
	}

	// Google's own docs specify this as a real JSON boolean (unlike Apple,
	// which has shipped both a bool and a "true"/"false" string across SDK
	// versions) -- but tolerating both shapes here costs nothing and
	// removes any risk of a documented-but-unenforced claim shape silently
	// locking out verified users. See parseVerifiedFlag in apple.go.
	if !parseVerifiedFlag(claims["email_verified"]) {
		return "", ErrEmailNotVerified
	}

	email, _ = claims["email"].(string)
	if email == "" {
		return "", errors.New("oauthverify: google id_token has no email claim")
	}

	return email, nil
}
