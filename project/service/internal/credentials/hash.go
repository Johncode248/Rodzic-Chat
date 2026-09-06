package credentials

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashToken returns the sha256 hex digest of a token string. Refresh tokens
// and password-reset tokens are stored (and looked up) by this hash, never
// by their raw value, so a DB read alone can't be replayed as a credential.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
