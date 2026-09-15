package oauthverify

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// newMockJWKSServer starts an httptest server publishing key's public half
// as a single-entry JWKS under kid, mirroring exactly what
// accounts.google.com/appleid.apple.com serve.
func newMockJWKSServer(t *testing.T, kid string, key *rsa.PrivateKey) *httptest.Server {
	t.Helper()

	n := base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.PublicKey.E)).Bytes())

	set := jwkSet{Keys: []jwk{{Kty: "RSA", Kid: kid, N: n, E: e}}}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(set)
	}))
}

func signToken(t *testing.T, key *rsa.PrivateKey, kid string, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("signing test token: %v", err)
	}
	return signed
}

// TestVerify_ValidTokenAccepted proves the positive path: a token
// genuinely signed by the key published at the JWKS URL, with the right
// issuer/audience/email_verified, is accepted and its email extracted --
// exercising the exact same jwt.ParseWithClaims + keyCache.publicKey path
// VerifyGoogleIDToken/VerifyAppleIdentityToken use, just against a local
// mock server instead of the real Google/Apple endpoints.
func TestVerify_ValidTokenAccepted(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	server := newMockJWKSServer(t, "test-kid", key)
	defer server.Close()

	cache := newKeyCache(server.URL)

	signed := signToken(t, key, "test-kid", jwt.MapClaims{
		"iss":            "https://accounts.google.com",
		"aud":            "test-audience",
		"email":          "parent@example.com",
		"email_verified": true,
		"exp":            time.Now().Add(time.Hour).Unix(),
	})

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(signed, claims, func(t *jwt.Token) (interface{}, error) {
		kid, _ := t.Header["kid"].(string)
		return cache.publicKey(kid)
	})
	if err != nil || !token.Valid {
		t.Fatalf("expected valid token, got err=%v valid=%v", err, token.Valid)
	}
	if email, _ := claims["email"].(string); email != "parent@example.com" {
		t.Fatalf("expected email parent@example.com, got %q", email)
	}
}

// TestVerify_WrongSigningKeyRejected proves a token claiming the right
// issuer/audience/email but signed by a DIFFERENT key than the one
// published at the JWKS URL is rejected -- the actual attack this whole
// package exists to stop (forging "I am this email" without controlling
// the provider account).
func TestVerify_WrongSigningKeyRejected(t *testing.T) {
	realKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	attackerKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	server := newMockJWKSServer(t, "real-kid", realKey)
	defer server.Close()
	cache := newKeyCache(server.URL)

	forged := signToken(t, attackerKey, "real-kid", jwt.MapClaims{
		"iss":            "https://accounts.google.com",
		"aud":            "test-audience",
		"email":          "victim@example.com",
		"email_verified": true,
		"exp":            time.Now().Add(time.Hour).Unix(),
	})

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(forged, claims, func(t *jwt.Token) (interface{}, error) {
		kid, _ := t.Header["kid"].(string)
		return cache.publicKey(kid)
	})
	if err == nil && token.Valid {
		t.Fatal("expected forged token (signed by wrong key) to be rejected, but it was accepted")
	}
}
