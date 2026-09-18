package parentchildpg

import (
	"time"

	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
)

type ParentChild struct {
	Id         string              `json:"id,omitempty" gorm:"primaryKey"`
	ParentId   string              `json:"parent_id,omitempty" gorm:"not null;index"`
	Parent     *accountspg.Account `json:"-" gorm:"foreignKey:ParentId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	ChildId    string              `json:"child_id,omitempty" gorm:"not null;index"`
	Child      *childrenpg.Child   `json:"-" gorm:"foreignKey:ChildId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Relation   RelationType        `json:"relation,omitempty" gorm:"not null"`
	CreateTime time.Time           `json:"create_time,omitempty" gorm:"not null"`
}
