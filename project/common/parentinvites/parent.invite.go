package parentinvitespg

import "time"

// ParentInvite lets one parent ("rodzic nr 1") invite another
// ("rodzic nr 2" -- a co-parent/spouse, typically) to manage the exact
// same children -- entirely peer-to-peer, no organization/director
// involved (unlike project/common/registrationlinks, which is a
// director's placowka invite). Accepting one links the accepting account
// to every one of the inviter's children via project/common/parentchild,
// same as any other parent link. Single-use (ConsumedByAccountId sits
// empty until someone accepts), expires after 7 days.
type ParentInvite struct {
	Id                  string     `json:"id,omitempty" gorm:"primaryKey"`
	Token               string     `json:"token,omitempty" gorm:"not null;uniqueIndex"`
	InviterAccountId    string     `json:"inviter_account_id,omitempty" gorm:"not null;index"`
	ExpiresAt           time.Time  `json:"expires_at,omitempty" gorm:"not null"`
	ConsumedByAccountId string     `json:"consumed_by_account_id,omitempty"`
	ConsumedAt          *time.Time `json:"consumed_at,omitempty"`
	CreateTime          time.Time  `json:"create_time,omitempty" gorm:"not null"`
}

// IsUsable reports whether the invite can still be accepted.
func (i ParentInvite) IsUsable(now time.Time) bool {
	return i.ConsumedByAccountId == "" && now.Before(i.ExpiresAt)
}
