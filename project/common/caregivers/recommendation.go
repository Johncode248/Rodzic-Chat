package caregiverspg

import (
	"time"

	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
)

// CaregiverRecommendation is a parent's "Polecam" endorsement of a
// Caregiver they aren't themselves -- unique per (caregiver_id,
// account_id): recommending again just updates the comment rather than
// creating a duplicate row, same convention as chats.MessageReaction.
type CaregiverRecommendation struct {
	Id          string              `json:"id,omitempty" gorm:"primaryKey"`
	CaregiverId string              `json:"caregiver_id,omitempty" gorm:"not null;index;uniqueIndex:idx_caregiver_recommendation"`
	Caregiver   *Caregiver          `json:"-" gorm:"foreignKey:CaregiverId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	AccountId   string              `json:"account_id,omitempty" gorm:"not null;index;uniqueIndex:idx_caregiver_recommendation"`
	Account     *accountspg.Account `json:"-" gorm:"foreignKey:AccountId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Comment     string              `json:"comment,omitempty"`
	CreateTime  time.Time           `json:"create_time,omitempty" gorm:"not null"`
}
