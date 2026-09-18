package credentialspg

import (
	"time"
)

type Credential struct {
	// UserId is the primary key (not a separate Id column): a Credential is
	// strictly 1:1 with an Account, looked up by user_id everywhere in this
	// package. No FK association to accountspg.Account here -- accountspg
	// already imports this package (for its own delete-condition plumbing),
	// so declaring one here would create an import cycle; the relationship
	// still exists, it's just not DB-enforced.
	UserId     string    `json:"user_id,omitempty" gorm:"primaryKey"`
	Password   string    `json:"password,omitempty" gorm:"not null"`
	CreateTime time.Time `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime time.Time `json:"update_time,omitempty" gorm:"not null"`
}
