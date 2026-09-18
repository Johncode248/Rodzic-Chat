package chatspg

import (
	"time"

	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
)

// MessageReaction is a lightweight emoji ack so parents don't have to reply
// "dziękuję" to every announcement, cluttering the thread. Unique per
// (message_id, account_id): posting a new reaction replaces the old one.
type MessageReaction struct {
	Id         string              `json:"id,omitempty" gorm:"primaryKey"`
	MessageId  string              `json:"message_id,omitempty" gorm:"not null;index;uniqueIndex:idx_message_reaction"`
	Message    *Message            `json:"-" gorm:"foreignKey:MessageId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	AccountId  string              `json:"account_id,omitempty" gorm:"not null;index;uniqueIndex:idx_message_reaction"`
	Account    *accountspg.Account `json:"-" gorm:"foreignKey:AccountId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Emoji      string              `json:"emoji,omitempty" gorm:"not null"`
	CreateTime time.Time           `json:"create_time,omitempty" gorm:"not null"`
}
