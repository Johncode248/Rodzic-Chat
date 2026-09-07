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
	// ChildrenLimit is the capacity tier the director picked at registration
	// (one of ChildrenLimitOptions(), see pricing.go), used both to compute
	// the monthly platform fee and to hard-block adding more children than
	// paid for (see project/service/internal/children.Create). Zero means
	// "not set" -- only possible for rows created before this field existed
	// -- and is treated as unlimited rather than zero capacity.
	ChildrenLimit int `json:"children_limit,omitempty" gorm:"not null;default:0"`
	// PaymentReference is a short, stable, generated code (e.g. "RC-7F3K9Q")
	// the director puts in the bank transfer title when paying their
	// platform subscription by hand -- see § "Rozliczenie platformy" in the
	// web app. Lets the operator match an incoming transfer to this
	// organization even if the sender's name on the transfer differs.
	PaymentReference string `json:"payment_reference,omitempty" gorm:"unique"`
	// SubscriptionFreeOverride, when true, exempts this organization from
	// the platform subscription fee entirely. Settable only by the
	// platform operator (see transport/organizations/subscription.override.go,
	// gated the same way as project/service/transport/platformsettings) --
	// never by the director themselves.
	SubscriptionFreeOverride bool      `json:"subscription_free_override,omitempty"`
	CreateTime               time.Time `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime               time.Time `json:"update_time,omitempty" gorm:"not null"`
}
