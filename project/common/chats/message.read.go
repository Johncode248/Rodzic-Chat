package chatspg

import (
	"time"

	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
)

// MessageRead is a read-receipt, reused for both chat messages and
// announcements (announcements are just messages posted into an
// Announcement-type channel). Unique per (message_id, account_id): marking
// read again is idempotent and just bumps ReadAt.
type MessageRead struct {
	Id        string              `json:"id,omitempty" gorm:"primaryKey"`
	MessageId string              `json:"message_id,omitempty" gorm:"not null;index;uniqueIndex:idx_message_read"`
	Message   *Message            `json:"-" gorm:"foreignKey:MessageId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	AccountId string              `json:"account_id,omitempty" gorm:"not null;index;uniqueIndex:idx_message_read"`
	Account   *accountspg.Account `json:"-" gorm:"foreignKey:AccountId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	ReadAt    time.Time           `json:"read_at,omitempty" gorm:"not null"`
}
