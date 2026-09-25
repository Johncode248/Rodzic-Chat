package chatspg

import (
	"time"

	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
)

type Poll struct {
	Id            string     `json:"id,omitempty" gorm:"primaryKey"`
	MessageId     string     `json:"message_id,omitempty" gorm:"not null;index"`
	Message       *Message   `json:"-" gorm:"foreignKey:MessageId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Question      string     `json:"question,omitempty" gorm:"not null"`
	AllowMultiple bool       `json:"allow_multiple,omitempty"`
	ClosesAt      *time.Time `json:"closes_at,omitempty"`
	CreateTime    time.Time  `json:"create_time,omitempty" gorm:"not null"`
	// Anonymous hides who voted for what (only tallies are returned). Kind is "" for a normal
	// choice poll or "dates" for a date/time poll whose option texts are RFC3339 timestamps.
	Anonymous bool   `json:"anonymous,omitempty"`
	Kind      string `json:"kind,omitempty"`
}

type PollOption struct {
	Id         string `json:"id,omitempty" gorm:"primaryKey"`
	PollId     string `json:"poll_id,omitempty" gorm:"not null;index"`
	Poll       *Poll  `json:"-" gorm:"foreignKey:PollId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Text       string `json:"text,omitempty" gorm:"not null"`
	OrderIndex int    `json:"order_index,omitempty"`
}

type PollVote struct {
	Id     string `json:"id,omitempty" gorm:"primaryKey"`
	PollId string `json:"poll_id,omitempty" gorm:"not null;index"`
	// The uniqueIndex below is defense-in-depth on top of the
	// pg_advisory_xact_lock in chats.CreateVote: it guarantees at the DB
	// level that one account can never hold two rows for the same
	// (poll, option) pair, which is exactly what let a multi-choice poll's
	// vote count be inflated by repeated calls before this existed.
	OptionId   string              `json:"option_id,omitempty" gorm:"not null;uniqueIndex:idx_poll_vote_unique"`
	Option     *PollOption         `json:"-" gorm:"foreignKey:OptionId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	AccountId  string              `json:"account_id,omitempty" gorm:"not null;uniqueIndex:idx_poll_vote_unique"`
	Account    *accountspg.Account `json:"-" gorm:"foreignKey:AccountId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	CreateTime time.Time           `json:"create_time,omitempty" gorm:"not null"`
}
