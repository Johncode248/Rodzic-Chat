package chatspg

import (
	"time"

	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
)

// ChannelMember is an explicit per-channel membership list. It exists only
// for org-less channels -- a ChannelTypeDirect channel created via a
// peer-to-peer chat invite (see project/common/chatinvites), so two parents
// with no shared placówka can still chat -- since every organization-scoped
// channel stays authorized by OrganizationId alone, exactly as before this
// table existed.
type ChannelMember struct {
	Id         string              `json:"id,omitempty" gorm:"primaryKey"`
	ChannelId  string              `json:"channel_id,omitempty" gorm:"not null;index;uniqueIndex:idx_channel_member"`
	Channel    *Channel            `json:"-" gorm:"foreignKey:ChannelId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	AccountId  string              `json:"account_id,omitempty" gorm:"not null;index;uniqueIndex:idx_channel_member"`
	Account    *accountspg.Account `json:"-" gorm:"foreignKey:AccountId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Role       ChannelMemberRole   `json:"role,omitempty" gorm:"not null"`
	CreateTime time.Time           `json:"create_time,omitempty" gorm:"not null"`
}
