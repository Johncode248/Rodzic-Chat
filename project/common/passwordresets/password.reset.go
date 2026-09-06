package passwordresetspg

import (
	"time"
)

// PasswordReset backs a single-use, short-lived password-reset link. The
// token embedded in the emailed link is random and opaque (library/random),
// never a JWT -- only its sha256 hash is stored, so a DB read alone can't be
// replayed, and consumption is tracked explicitly (UsedAt) so the same link
// can't be used twice.
type PasswordReset struct {
	Id         string     `json:"id,omitempty" gorm:"primaryKey"`
	AccountId  string     `json:"account_id,omitempty" gorm:"not null;index"`
	TokenHash  string     `json:"-" gorm:"not null;index"`
	ExpiresAt  time.Time  `json:"expires_at,omitempty" gorm:"not null"`
	UsedAt     *time.Time `json:"used_at,omitempty"`
	CreateTime time.Time  `json:"create_time,omitempty" gorm:"not null"`
}
