package parentchildpg

import (
	"time"
)

type ParentChild struct {
	Id         string       `json:"id,omitempty" gorm:"primaryKey"`
	ParentId   string       `json:"parent_id,omitempty" gorm:"not null;index"`
	ChildId    string       `json:"child_id,omitempty" gorm:"not null;index"`
	Relation   RelationType `json:"relation,omitempty" gorm:"not null"`
	CanPickup  bool         `json:"can_pickup,omitempty"`
	CreateTime time.Time    `json:"create_time,omitempty" gorm:"not null"`
}
