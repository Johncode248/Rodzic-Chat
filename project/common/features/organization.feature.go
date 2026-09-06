package featurespg

import (
	"time"
)

type OrganizationFeature struct {
	Id             string    `json:"id,omitempty" gorm:"primaryKey"`
	OrganizationId string    `json:"organization_id,omitempty" gorm:"not null;index;uniqueIndex:idx_org_feature"`
	Feature        Feature   `json:"feature,omitempty" gorm:"not null;uniqueIndex:idx_org_feature"`
	Enabled        bool      `json:"enabled" gorm:"not null"`
	CreateTime     time.Time `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime     time.Time `json:"update_time,omitempty" gorm:"not null"`
}
