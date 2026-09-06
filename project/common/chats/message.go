package chatspg

import "time"

type Message struct {
	Id             string      `json:"id,omitempty" gorm:"primaryKey"`
	ChannelId      string      `json:"channel_id,omitempty" gorm:"not null;index"`
	OrganizationId string      `json:"organization_id,omitempty" gorm:"not null;index"`
	SenderId       string      `json:"sender_id,omitempty" gorm:"not null"`
	Type           MessageType `json:"type,omitempty" gorm:"not null"`
	Content        string      `json:"content,omitempty"`
	AttachmentUrl  string      `json:"attachment_url,omitempty"`
	CreateTime     time.Time   `json:"create_time,omitempty" gorm:"not null"`
}
