package placespg

import (
	"time"

	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
)

// Recommendation is one parent's "poleca to miejsce" -- at most one per
// (PlaceId, AccountId) pair (enforced by a unique index; see
// project/service/internal/places/recommend.go, which upserts rather than
// erroring on a repeat recommend). A place's total recommendation count is
// simply the row count for its PlaceId -- computed on read, not cached on
// Place, so it's never at risk of drifting.
type Recommendation struct {
	Id         string              `json:"id,omitempty" gorm:"primaryKey"`
	PlaceId    string              `json:"place_id,omitempty" gorm:"uniqueIndex:idx_place_account;not null"`
	Place      *Place              `json:"-" gorm:"foreignKey:PlaceId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	AccountId  string              `json:"account_id,omitempty" gorm:"uniqueIndex:idx_place_account;not null"`
	Account    *accountspg.Account `json:"-" gorm:"foreignKey:AccountId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Comment    string              `json:"comment,omitempty"`
	CreateTime time.Time           `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime time.Time           `json:"update_time,omitempty" gorm:"not null"`
}
