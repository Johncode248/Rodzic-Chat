package chatspg

import "time"

// MessageReaction is a lightweight emoji ack so parents don't have to reply
// "dziękuję" to every announcement, cluttering the thread. Unique per
// (message_id, account_id): posting a new reaction replaces the old one.
type MessageReaction struct {
	Id         string    `json:"id,omitempty" gorm:"primaryKey"`
	MessageId  string    `json:"message_id,omitempty" gorm:"not null;index;uniqueIndex:idx_message_reaction"`
	AccountId  string    `json:"account_id,omitempty" gorm:"not null;index;uniqueIndex:idx_message_reaction"`
	Emoji      string    `json:"emoji,omitempty" gorm:"not null"`
	CreateTime time.Time `json:"create_time,omitempty" gorm:"not null"`
}
