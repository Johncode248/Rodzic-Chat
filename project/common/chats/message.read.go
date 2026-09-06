package chatspg

import "time"

// MessageRead is a read-receipt, reused for both chat messages and
// announcements (announcements are just messages posted into an
// Announcement-type channel). Unique per (message_id, account_id): marking
// read again is idempotent and just bumps ReadAt.
type MessageRead struct {
	Id        string    `json:"id,omitempty" gorm:"primaryKey"`
	MessageId string    `json:"message_id,omitempty" gorm:"not null;index;uniqueIndex:idx_message_read"`
	AccountId string    `json:"account_id,omitempty" gorm:"not null;index;uniqueIndex:idx_message_read"`
	ReadAt    time.Time `json:"read_at,omitempty" gorm:"not null"`
}
