package credentialspg

import (
	"time"
)

type Credential struct {
	UserId     string    `json:"user_id,omitempty" gorm:"not null"`
	Password   string    `json:"password,omitempty" gorm:"not null"`
	CreateTime time.Time `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime time.Time `json:"update_time,omitempty" gorm:"not null"`
}
