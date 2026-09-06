package credentials

import "time"

// AccessTokenTTL/RefreshTokenTTL are the single source of truth for token
// lifetimes -- generation and every reply's ExpiresAt/ExpiresIn must derive
// from these rather than hardcoding the duration again, which is what
// previously let LoginReply (23h) drift out of sync with the access token
// itself (24h).
const (
	AccessTokenTTL  = 24 * time.Hour
	RefreshTokenTTL = 720 * time.Hour // 30 days
)
