package chatspg

import (
	"time"

	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
)

type Message struct {
	Id             string                        `json:"id,omitempty" gorm:"primaryKey"`
	ChannelId      string                        `json:"channel_id,omitempty" gorm:"not null;index"`
	Channel        *Channel                      `json:"-" gorm:"foreignKey:ChannelId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	OrganizationId string                        `json:"organization_id,omitempty" gorm:"not null;index"`
	Organization   *organizationspg.Organization `json:"-" gorm:"foreignKey:OrganizationId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	SenderId       string                        `json:"sender_id,omitempty" gorm:"not null;index"`
	Type           MessageType                   `json:"type,omitempty" gorm:"not null"`
	Content        string                        `json:"content,omitempty"`
	AttachmentUrl  string                        `json:"attachment_url,omitempty"`
	CreateTime     time.Time                     `json:"create_time,omitempty" gorm:"not null"`
	// SenderName is computed, never persisted -- see chats.AttachSenderNames. Empty for a system
	// message (SenderId == "") or when the caller's list endpoint didn't populate it.
	SenderName string `json:"sender_name,omitempty" gorm:"-"`
}
