package chatinvitespg

import (
	"time"

	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
)

// ChatInvite lets one parent invite another -- no shared placówka required
// -- into a private 1:1 chat. Entirely peer-to-peer, mirrors
// project/common/parentinvites.ParentInvite but accepting one adds a
// chatspg.ChannelMember row instead of linking children. Single-use
// (ConsumedByAccountId sits empty until someone accepts), expires after 72
// hours. Both inviter and accepter must already be authenticated accounts
// -- PinHash is the sha256 digest (see library/pinhash) of an 8-digit PIN
// (see library/random.GeneratePin); the plaintext PIN is returned once, on
// creation, and never stored anywhere.
type ChatInvite struct {
	Id                  string           `json:"id,omitempty" gorm:"primaryKey"`
	PinHash             string           `json:"-" gorm:"not null;uniqueIndex"`
	ChannelId           string           `json:"channel_id,omitempty" gorm:"not null;index"`
	Channel             *chatspg.Channel `json:"-" gorm:"foreignKey:ChannelId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	InviterAccountId    string           `json:"inviter_account_id,omitempty" gorm:"not null;index"`
	ExpiresAt           time.Time        `json:"expires_at,omitempty" gorm:"not null"`
	ConsumedByAccountId string           `json:"consumed_by_account_id,omitempty"`
	ConsumedAt          *time.Time       `json:"consumed_at,omitempty"`
	CreateTime          time.Time        `json:"create_time,omitempty" gorm:"not null"`
}

// IsUsable reports whether the invite can still be accepted.
func (i ChatInvite) IsUsable(now time.Time) bool {
	return i.ConsumedByAccountId == "" && now.Before(i.ExpiresAt)
}
