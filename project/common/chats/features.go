package chatspg

import (
	"time"

	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
)

// PersonRef is a display-ready reference to an account (name resolved server-side), optionally
// with the time of the action it is listed for (a confirmation, a claim).
type PersonRef struct {
	AccountId string     `json:"account_id,omitempty"`
	Name      string     `json:"name,omitempty"`
	Time      *time.Time `json:"time,omitempty"`
}

// ---- Consents ("Potwierdź, że przeczytałeś/zgadzasz się") --------------------------------------

// MessageConfirmation records that an account confirmed a Consent-type message. Unique per
// (message, account): confirming twice is idempotent.
type MessageConfirmation struct {
	Id          string              `json:"id,omitempty" gorm:"primaryKey"`
	MessageId   string              `json:"message_id,omitempty" gorm:"not null;index;uniqueIndex:idx_message_confirmation"`
	Message     *Message            `json:"-" gorm:"foreignKey:MessageId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	AccountId   string              `json:"account_id,omitempty" gorm:"not null;index;uniqueIndex:idx_message_confirmation"`
	Account     *accountspg.Account `json:"-" gorm:"foreignKey:AccountId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	ConfirmTime time.Time           `json:"confirm_time,omitempty" gorm:"not null"`
}

type ConsentCreateRequest struct {
	ChannelId string `json:"channel_id,omitempty"`
	Content   string `json:"content,omitempty"`
}

type ConsentCreateResponse struct {
	Message Message `json:"message,omitempty"`
}

type ConsentMessageRequest struct {
	MessageId string `json:"message_id,omitempty"`
}

// ConsentStatusResponse is the caller's view of a consent: everyone sees the counts and whether
// they themselves confirmed; only staff also get the named Confirmed/Pending lists.
type ConsentStatusResponse struct {
	MessageId      string      `json:"message_id,omitempty"`
	ConfirmedCount int         `json:"confirmed_count"`
	TotalCount     int         `json:"total_count"`
	IConfirmed     bool        `json:"i_confirmed"`
	Confirmed      []PersonRef `json:"confirmed,omitempty"`
	Pending        []PersonRef `json:"pending,omitempty"`
}

type ConsentRemindResponse struct {
	Reminded int `json:"reminded"`
}

// ---- Signup lists ("Kto przyniesie...") --------------------------------------------------------

// SignupSlot is one line of a signup list; Capacity 0 means unlimited people.
type SignupSlot struct {
	Id         string   `json:"id,omitempty" gorm:"primaryKey"`
	MessageId  string   `json:"message_id,omitempty" gorm:"not null;index"`
	Message    *Message `json:"-" gorm:"foreignKey:MessageId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Title      string   `json:"title,omitempty" gorm:"not null"`
	Capacity   int      `json:"capacity,omitempty"`
	OrderIndex int      `json:"order_index,omitempty"`
}

// SignupClaim is one account taking one slot. Unique per (slot, account).
type SignupClaim struct {
	Id         string              `json:"id,omitempty" gorm:"primaryKey"`
	SlotId     string              `json:"slot_id,omitempty" gorm:"not null;index;uniqueIndex:idx_signup_claim"`
	Slot       *SignupSlot         `json:"-" gorm:"foreignKey:SlotId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	AccountId  string              `json:"account_id,omitempty" gorm:"not null;index;uniqueIndex:idx_signup_claim"`
	Account    *accountspg.Account `json:"-" gorm:"foreignKey:AccountId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	CreateTime time.Time           `json:"create_time,omitempty" gorm:"not null"`
}

type SignupSlotInput struct {
	Title    string `json:"title,omitempty"`
	Capacity int    `json:"capacity,omitempty"`
}

type SignupCreateRequest struct {
	ChannelId string            `json:"channel_id,omitempty"`
	Title     string            `json:"title,omitempty"`
	Slots     []SignupSlotInput `json:"slots,omitempty"`
}

type SignupSlotDetail struct {
	Slot     SignupSlot  `json:"slot,omitempty"`
	Claimed  []PersonRef `json:"claimed,omitempty"`
	IClaimed bool        `json:"i_claimed"`
	Full     bool        `json:"full"`
}

type SignupDetailResponse struct {
	MessageId string             `json:"message_id,omitempty"`
	Title     string             `json:"title,omitempty"`
	Slots     []SignupSlotDetail `json:"slots,omitempty"`
}

type SignupCreateResponse struct {
	Message Message              `json:"message,omitempty"`
	Detail  SignupDetailResponse `json:"detail,omitempty"`
}

type SignupClaimRequest struct {
	SlotId string `json:"slot_id,omitempty"`
}
