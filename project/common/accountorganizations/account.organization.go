package accountorganizationspg

import (
	"time"

	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
)

// AccountOrganization is a parent account's membership in one facility
// (placówka). A parent can belong to several -- being removed from one
// (an AccountOrganization row deleted) doesn't touch the others. Directors
// and teachers stay effectively single-org via Account.OrganizationId (see
// account.go's comment); this table's real purpose is parents with children
// at more than one facility.
type AccountOrganization struct {
	Id             string                        `json:"id,omitempty" gorm:"primaryKey"`
	AccountId      string                        `json:"account_id,omitempty" gorm:"not null;index;uniqueIndex:idx_account_org"`
	Account        *accountspg.Account           `json:"-" gorm:"foreignKey:AccountId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	OrganizationId string                        `json:"organization_id,omitempty" gorm:"not null;index;uniqueIndex:idx_account_org"`
	Organization   *organizationspg.Organization `json:"-" gorm:"foreignKey:OrganizationId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	IsPrimary      bool                          `json:"is_primary,omitempty"`
	CreateTime     time.Time                     `json:"create_time,omitempty" gorm:"not null"`
}
