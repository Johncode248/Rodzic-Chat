package chatspg

import "time"

type Channel struct {
	Id             string      `json:"id,omitempty" gorm:"primaryKey"`
	OrganizationId string      `json:"organization_id,omitempty" gorm:"not null;index"`
	GroupId        string      `json:"group_id,omitempty" gorm:"index"`
	Type           ChannelType `json:"type,omitempty" gorm:"not null"`
	Name           string      `json:"name,omitempty" gorm:"not null"`
	CreateTime     time.Time   `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime     time.Time   `json:"update_time,omitempty" gorm:"not null"`
}
