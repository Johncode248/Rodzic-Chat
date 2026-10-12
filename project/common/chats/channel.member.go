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
//
// Role is the preset (member/moderator/admin); Permissions are extra bits granted to this one
// member on top of it -- see ChannelPermission and EffectivePermissions.
type ChannelMember struct {
	Id          string              `json:"id,omitempty" gorm:"primaryKey"`
	ChannelId   string              `json:"channel_id,omitempty" gorm:"not null;index;uniqueIndex:idx_channel_member"`
	Channel     *Channel            `json:"-" gorm:"foreignKey:ChannelId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	AccountId   string              `json:"account_id,omitempty" gorm:"not null;index;uniqueIndex:idx_channel_member"`
	Account     *accountspg.Account `json:"-" gorm:"foreignKey:AccountId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Role        ChannelMemberRole   `json:"role,omitempty" gorm:"not null"`
	Permissions ChannelPermission   `json:"permissions,omitempty" gorm:"not null;default:0"`
	CreateTime  time.Time           `json:"create_time,omitempty" gorm:"not null"`
}

// EffectivePermissions is the role preset OR-ed with this member's individual grants.
func (m ChannelMember) EffectivePermissions() ChannelPermission {
	return RolePermissions(m.Role) | m.Permissions
}

// ChannelBan blocks one member of a channel from reading/posting until Until (nil = until lifted).
// The membership itself is kept, so when a temporary ban ends the person is simply back.
type ChannelBan struct {
	Id         string              `json:"id,omitempty" gorm:"primaryKey"`
	ChannelId  string              `json:"channel_id,omitempty" gorm:"not null;index;index:idx_channel_ban,priority:1"`
	Channel    *Channel            `json:"-" gorm:"foreignKey:ChannelId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	AccountId  string              `json:"account_id,omitempty" gorm:"not null;index;index:idx_channel_ban,priority:2"`
	Account    *accountspg.Account `json:"-" gorm:"foreignKey:AccountId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	BannedBy   string              `json:"banned_by,omitempty" gorm:"not null"`
	Reason     string              `json:"reason,omitempty"`
	Until      *time.Time          `json:"until,omitempty"`
	CreateTime time.Time           `json:"create_time,omitempty" gorm:"not null"`
}

// Active reports whether the ban is still in force at t.
func (b ChannelBan) Active(t time.Time) bool { return b.Until == nil || b.Until.After(t) }
