package chatspg

import (
	"time"
)

type Message struct {
	Id        string   `json:"id,omitempty" gorm:"primaryKey"`
	ChannelId string   `json:"channel_id,omitempty" gorm:"not null;index"`
	Channel   *Channel `json:"-" gorm:"foreignKey:ChannelId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	// No FK to organizations: org-less parent-to-parent channels use OrganizationId "".
	OrganizationId string      `json:"organization_id,omitempty" gorm:"not null;index"`
	SenderId       string      `json:"sender_id,omitempty" gorm:"not null;index"`
	Type           MessageType `json:"type,omitempty" gorm:"not null"`
	Content        string      `json:"content,omitempty"`
	AttachmentUrl  string      `json:"attachment_url,omitempty"`
	CreateTime     time.Time   `json:"create_time,omitempty" gorm:"not null"`
	// Urgent marks a message the sender wants noticed (shown highlighted, pushed with a "Pilne"
	// prefix). PinnedAt is non-nil while a message is pinned to the top of its channel.
	Urgent   bool       `json:"urgent,omitempty"`
	PinnedAt *time.Time `json:"pinned_at,omitempty"`
	PinnedBy string     `json:"pinned_by,omitempty"`
	// ReplyToId is the message this one answers (same channel); empty for a top-level message.
	ReplyToId string `json:"reply_to_id,omitempty"`
	// SenderName is computed, never persisted -- see chats.AttachSenderNames. Empty for a system
	// message (SenderId == "") or when the caller's list endpoint didn't populate it.
	SenderName string `json:"sender_name,omitempty" gorm:"-"`
}
