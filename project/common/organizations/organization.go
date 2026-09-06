package organizationspg

import (
	"time"
)

type Organization struct {
	Id                 string             `json:"id,omitempty" gorm:"primaryKey"`
	Name               string             `json:"name,omitempty" gorm:"not null"`
	Type               OrganizationType   `json:"type,omitempty" gorm:"not null"`
	Subdomain          string             `json:"subdomain,omitempty" gorm:"unique; not null"`
	Address            string             `json:"address,omitempty"`
	Nip                string             `json:"nip,omitempty"`
	LogoUrl            string             `json:"logo_url,omitempty"`
	PrimaryColor       string             `json:"primary_color,omitempty"`
	SubscriptionStatus SubscriptionStatus `json:"subscription_status,omitempty" gorm:"not null"`
	SubscriptionPlan   string             `json:"subscription_plan,omitempty"`
	CreateTime         time.Time          `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime         time.Time          `json:"update_time,omitempty" gorm:"not null"`
}
