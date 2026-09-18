package devicetokenspg

import (
	"time"

	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
)

// Platform is which push service a token belongs to -- Android and iOS
// both go through Firebase Cloud Messaging, but the payload shape FCM
// needs to build differs slightly per platform (e.g. APNs-specific fields
// for iOS), so it's still worth recording.
type Platform string

const (
	PlatformAndroid Platform = "android"
	PlatformIOS     Platform = "ios"
)

func (p Platform) Valid() bool {
	return p == PlatformAndroid || p == PlatformIOS
}

// DeviceToken is one (account, physical device install) pair's current FCM
// registration token. Unique on Token, not on (AccountId, Token): the same
// physical device can only ever hold one live FCM token at a time (a fresh
// install/reinstall gets a new one), but logging out and back in as a
// different account on that device must repoint the existing row at the
// new account rather than erroring or duplicating -- see devicetokens.Register.
type DeviceToken struct {
	Id         string              `json:"id,omitempty" gorm:"primaryKey"`
	AccountId  string              `json:"account_id,omitempty" gorm:"not null;index"`
	Account    *accountspg.Account `json:"-" gorm:"foreignKey:AccountId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Platform   Platform            `json:"platform,omitempty" gorm:"not null"`
	Token      string              `json:"-" gorm:"not null;uniqueIndex"`
	CreateTime time.Time           `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime time.Time           `json:"update_time,omitempty" gorm:"not null"`
}

// RegisterRequest is what POST /private/device-tokens takes: the caller's
// FCM registration token for this install, refreshed every app launch (FCM
// tokens can rotate) and right after login.
type RegisterRequest struct {
	Token    string `json:"token,omitempty"`
	Platform string `json:"platform,omitempty"`
}

// UnregisterRequest is what DELETE /private/device-tokens takes -- sent on
// logout so a signed-out device stops receiving pushes meant for the
// account that just left it.
type UnregisterRequest struct {
	Token string `json:"token,omitempty"`
}
