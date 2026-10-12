package avatarspg

import (
	"time"

	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
)

// MaxBytes caps an uploaded avatar; clients downscale to ~256 px JPEG first (a few tens of KB).
const MaxBytes = 512 * 1024

// Avatar is an account's profile picture, stored as the (already downscaled) image bytes themselves --
// small enough for the database, so no object storage is needed for it.
type Avatar struct {
	AccountId   string              `json:"account_id,omitempty" gorm:"primaryKey"`
	Account     *accountspg.Account `json:"-" gorm:"foreignKey:AccountId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	ContentType string              `json:"content_type,omitempty" gorm:"not null"`
	Data        []byte              `json:"-" gorm:"not null"`
	UpdateTime  time.Time           `json:"update_time,omitempty" gorm:"not null"`
}
