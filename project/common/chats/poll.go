package chatspg

import "time"

type Poll struct {
	Id            string     `json:"id,omitempty" gorm:"primaryKey"`
	MessageId     string     `json:"message_id,omitempty" gorm:"not null;index"`
	Question      string     `json:"question,omitempty" gorm:"not null"`
	AllowMultiple bool       `json:"allow_multiple,omitempty"`
	ClosesAt      *time.Time `json:"closes_at,omitempty"`
	CreateTime    time.Time  `json:"create_time,omitempty" gorm:"not null"`
}

type PollOption struct {
	Id         string `json:"id,omitempty" gorm:"primaryKey"`
	PollId     string `json:"poll_id,omitempty" gorm:"not null;index"`
	Text       string `json:"text,omitempty" gorm:"not null"`
	OrderIndex int    `json:"order_index,omitempty"`
}

type PollVote struct {
	Id         string    `json:"id,omitempty" gorm:"primaryKey"`
	PollId     string    `json:"poll_id,omitempty" gorm:"not null;index"`
	OptionId   string    `json:"option_id,omitempty" gorm:"not null;index"`
	AccountId  string    `json:"account_id,omitempty" gorm:"not null;index"`
	CreateTime time.Time `json:"create_time,omitempty" gorm:"not null"`
}
