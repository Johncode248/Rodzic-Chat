package todospg

import (
	"time"

	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
)

// Todo is a task a parent writes for themselves ("umówić wizytę", "kupić strój na przedstawienie").
// It lives next to the AUTOMATIC to-dos (unpaid payments, event sign-ups, polls/consents awaiting an
// answer), which are computed on the fly and have no row -- see internal/todos.List, which merges both.
type Todo struct {
	Id        string              `json:"id,omitempty" gorm:"primaryKey"`
	AccountId string              `json:"account_id,omitempty" gorm:"not null;index"`
	Account   *accountspg.Account `json:"-" gorm:"foreignKey:AccountId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Title     string              `json:"title,omitempty" gorm:"not null"`
	Note      string              `json:"note,omitempty"`
	DueDate   *time.Time          `json:"due_date,omitempty"`
	Done      bool                `json:"done" gorm:"not null;default:false"`
	DoneAt    *time.Time          `json:"done_at,omitempty"`
	// Phase B: who it concerns, importance, repetition, reminder, sharing with the other parent, checklist.
	ChildId    string     `json:"child_id,omitempty"`
	Important  bool       `json:"important" gorm:"not null;default:false"`
	RepeatRule string     `json:"repeat_rule,omitempty"` // "", daily, weekly, monthly, yearly
	RemindAt   *time.Time `json:"remind_at,omitempty"`
	Reminded   bool       `json:"-" gorm:"not null;default:false"`
	Shared     bool       `json:"shared" gorm:"not null;default:false"`
	Checklist  string     `json:"-"` // JSON [{"t":"...","d":false}]
	CreateTime time.Time  `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime time.Time  `json:"update_time,omitempty" gorm:"not null"`
}

// Sources of an Item in the merged list.
const (
	SourceOwn     = "own"
	SourcePayment = "payment"
	SourceEvent   = "event"
	SourcePoll    = "poll"
	SourceConsent = "consent"
)

// Item is one row of the merged Todo list. Automatic items (everything but SourceOwn) can't be
// ticked off -- they disappear once the underlying thing is done (paid, registered, answered).
// ChecklistEntry is one sub-step of a task.
type ChecklistEntry struct {
	Text string `json:"t"`
	Done bool   `json:"d"`
}

// Snooze hides an AUTOMATIC to-do (Key = its Item.Id, e.g. "payment:...") until Until.
type Snooze struct {
	AccountId string              `json:"account_id,omitempty" gorm:"primaryKey"`
	Account   *accountspg.Account `json:"-" gorm:"foreignKey:AccountId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Key       string              `json:"key,omitempty" gorm:"primaryKey"`
	Until     time.Time           `json:"until" gorm:"not null"`
}

func (Snooze) TableName() string { return "todo_snoozes" }
func (Prefs) TableName() string  { return "todo_prefs" }

// Prefs are the opt-in notification settings of the "Do zrobienia" screen (everything off by default).
type Prefs struct {
	AccountId       string              `json:"-" gorm:"primaryKey"`
	Account         *accountspg.Account `json:"-" gorm:"foreignKey:AccountId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Digest          bool                `json:"digest" gorm:"not null;default:false"`        // 8:00 "masz N zadań na dziś"
	DueReminders    bool                `json:"due_reminders" gorm:"not null;default:false"` // 18:00 "jutro: ..."
	LastDigestDay   string              `json:"-"`
	LastReminderDay string              `json:"-"`
}

type Item struct {
	Id       string     `json:"id,omitempty"`
	Source   string     `json:"source,omitempty"`
	Title    string     `json:"title,omitempty"`
	Subtitle string     `json:"subtitle,omitempty"`
	DueDate  *time.Time `json:"due_date,omitempty"`
	Done     bool       `json:"done"`
	// DoneAt is when an own task was ticked off (drives "zrobione dziś" on the Do zrobienia screen).
	DoneAt       *time.Time `json:"done_at,omitempty"`
	AmountGrosze int64      `json:"amount_grosze,omitempty"`
	Important    bool       `json:"important,omitempty"`
	RepeatRule   string     `json:"repeat_rule,omitempty"`
	RemindAt     *time.Time `json:"remind_at,omitempty"`
	// Shared: visible to / editable by the other parent. Mine: false for tasks the other parent shared with me.
	Shared    bool             `json:"shared,omitempty"`
	Mine      bool             `json:"mine"`
	OwnerName string           `json:"owner_name,omitempty"`
	Checklist []ChecklistEntry `json:"checklist,omitempty"`
	// Deep-link ids for the automatic kinds.
	ChannelId string `json:"channel_id,omitempty"`
	MessageId string `json:"message_id,omitempty"`
	EventId   string `json:"event_id,omitempty"`
	ChildId   string `json:"child_id,omitempty"`
	PaymentId string `json:"payment_id,omitempty"`
}

type ListResponse struct {
	Items []Item `json:"items"`
}

type CreateRequest struct {
	Title      string           `json:"title,omitempty"`
	Note       string           `json:"note,omitempty"`
	DueDate    string           `json:"due_date,omitempty"` // RFC3339, "" = no due date
	ChildId    string           `json:"child_id,omitempty"`
	Important  bool             `json:"important,omitempty"`
	RepeatRule string           `json:"repeat_rule,omitempty"`
	RemindAt   string           `json:"remind_at,omitempty"` // RFC3339, "" = none
	Shared     bool             `json:"shared,omitempty"`
	Checklist  []ChecklistEntry `json:"checklist,omitempty"`
}

type UpdateRequest struct {
	Title      *string           `json:"title,omitempty"`
	Note       *string           `json:"note,omitempty"`
	DueDate    *string           `json:"due_date,omitempty"` // "" clears
	Done       *bool             `json:"done,omitempty"`
	ChildId    *string           `json:"child_id,omitempty"`
	Important  *bool             `json:"important,omitempty"`
	RepeatRule *string           `json:"repeat_rule,omitempty"`
	RemindAt   *string           `json:"remind_at,omitempty"` // "" clears
	Shared     *bool             `json:"shared,omitempty"`
	Checklist  *[]ChecklistEntry `json:"checklist,omitempty"`
}

type SnoozeRequest struct {
	Key   string `json:"key"`
	Until string `json:"until"` // RFC3339
}
