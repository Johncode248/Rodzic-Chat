package sessionspg

import (
	"time"

	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
)

// Session is one logged-in device/refresh-token lineage for an Account. It
// backs persistent login: RefreshTokenHash lets /refresh-token validate and
// rotate in place (same row, new hash) instead of minting fully stateless
// tokens, and RevokedAt lets a single device be logged out (or an account's
// sessions be revoked wholesale on password reset / self-delete) without
// touching any other device's session.
type Session struct {
	Id               string              `json:"id,omitempty" gorm:"primaryKey"`
	AccountId        string              `json:"account_id,omitempty" gorm:"not null;index"`
	Account          *accountspg.Account `json:"-" gorm:"foreignKey:AccountId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	RefreshTokenHash string              `json:"-" gorm:"not null;index"`
	DeviceId         string              `json:"device_id,omitempty" gorm:"index"`
	DeviceName       string              `json:"device_name,omitempty"`
	UserAgent        string              `json:"user_agent,omitempty"`
	IpAddress        string              `json:"ip_address,omitempty"`
	IssuedAt         time.Time           `json:"issued_at,omitempty" gorm:"not null"`
	ExpiresAt        time.Time           `json:"expires_at,omitempty" gorm:"not null;index"`
	LastUsedAt       time.Time           `json:"last_used_at,omitempty"`
	RevokedAt        *time.Time          `json:"revoked_at,omitempty" gorm:"index"`
	CreateTime       time.Time           `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime       time.Time           `json:"update_time,omitempty" gorm:"not null"`
}
