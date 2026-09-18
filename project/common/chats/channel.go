package chatspg

import (
	"time"

	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
)

type Channel struct {
	Id             string                        `json:"id,omitempty" gorm:"primaryKey"`
	OrganizationId string                        `json:"organization_id,omitempty" gorm:"not null;index"`
	Organization   *organizationspg.Organization `json:"-" gorm:"foreignKey:OrganizationId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	GroupId        string                        `json:"group_id,omitempty" gorm:"index"`
	Type           ChannelType                   `json:"type,omitempty" gorm:"not null"`
	Name           string                        `json:"name,omitempty" gorm:"not null"`
	CreateTime     time.Time                     `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime     time.Time                     `json:"update_time,omitempty" gorm:"not null"`
	// Unread is computed, never persisted -- true when the channel's most recent message has no
	// MessageRead row for the requesting account (see chats.AttachUnreadFlags). Only populated by
	// the channel-list handlers, which know who's asking; zero-value (false) everywhere else.
	Unread bool `json:"unread,omitempty" gorm:"-"`
}
