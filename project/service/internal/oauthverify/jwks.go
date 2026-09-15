// Package oauthverify verifies the identity tokens native mobile sign-in
// SDKs hand back (Google Sign-In on Android, Sign in with Apple on iOS) --
// entirely server-side, without ever trusting the client's own claim of
// "this is my email". Both providers sign their tokens with a rotating RSA
// key published at a well-known JWKS URL; verifying the signature here is
// what makes the token's email claim trustworthy enough to look up or
// create an Account by.
package oauthverify

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"
)

type jwk struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwkSet struct {
	Keys []jwk `json:"keys"`
}

// keyCache fetches and caches one provider's JWKS, re-fetching only after
// ttl elapses -- both Google and Apple rotate these keys infrequently, and
// every token verification would otherwise mean an extra HTTP round trip.
type keyCache struct {
	url       string
	ttl       time.Duration
	mu        sync.Mutex
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time
}

func newKeyCache(url string) *keyCache {
	return &keyCache{url: url, ttl: time.Hour}
}

func (c *keyCache) publicKey(kid string) (*rsa.PublicKey, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.keys == nil || time.Since(c.fetchedAt) > c.ttl {
		fresh, err := fetchJWKS(c.url)
		if err != nil {
			if c.keys != nil {
				// Serve stale keys rather than fail outright if the
				// provider's endpoint is briefly unreachable.
				if key, ok := c.keys[kid]; ok {
					return key, nil
				}
			}
			return nil, err
		}
		c.keys = fresh
		c.fetchedAt = time.Now()
	}

	key, ok := c.keys[kid]
	if !ok {
		return nil, fmt.Errorf("oauthverify: no key found for kid %q", kid)
	}
	return key, nil
}

func fetchJWKS(url string) (map[string]*rsa.PublicKey, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oauthverify: fetching %s: unexpected status %d", url, resp.StatusCode)
	}

	var set jwkSet
	if err := json.NewDecoder(resp.Body).Decode(&set); err != nil {
		return nil, err
	}

	keys := make(map[string]*rsa.PublicKey, len(set.Keys))
	for _, k := range set.Keys {
		if k.Kty != "RSA" {
			continue
		}
		pub, err := rsaPublicKeyFromJWK(k)
		if err != nil {
			continue
		}
		keys[k.Kid] = pub
	}
	return keys, nil
}

func rsaPublicKeyFromJWK(k jwk) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{N: n, E: int(e.Int64())}, nil
}
