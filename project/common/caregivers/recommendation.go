package caregiverspg

import "time"

// CaregiverRecommendation is a parent's "Polecam" endorsement of a
// Caregiver they aren't themselves -- unique per (caregiver_id,
// account_id): recommending again just updates the comment rather than
// creating a duplicate row, same convention as chats.MessageReaction.
type CaregiverRecommendation struct {
	Id          string    `json:"id,omitempty" gorm:"primaryKey"`
	CaregiverId string    `json:"caregiver_id,omitempty" gorm:"not null;index;uniqueIndex:idx_caregiver_recommendation"`
	AccountId   string    `json:"account_id,omitempty" gorm:"not null;index;uniqueIndex:idx_caregiver_recommendation"`
	Comment     string    `json:"comment,omitempty"`
	CreateTime  time.Time `json:"create_time,omitempty" gorm:"not null"`
}
