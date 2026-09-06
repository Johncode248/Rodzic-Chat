package accountorganizationspg

import (
	"time"
)

// AccountOrganization is a parent account's membership in one facility
// (placówka). A parent can belong to several -- being removed from one
// (an AccountOrganization row deleted) doesn't touch the others. Directors
// and teachers stay effectively single-org via Account.OrganizationId (see
// account.go's comment); this table's real purpose is parents with children
// at more than one facility.
type AccountOrganization struct {
	Id             string    `json:"id,omitempty" gorm:"primaryKey"`
	AccountId      string    `json:"account_id,omitempty" gorm:"not null;index;uniqueIndex:idx_account_org"`
	OrganizationId string    `json:"organization_id,omitempty" gorm:"not null;index;uniqueIndex:idx_account_org"`
	IsPrimary      bool      `json:"is_primary,omitempty"`
	CreateTime     time.Time `json:"create_time,omitempty" gorm:"not null"`
}
