// Package shoppingpg holds the family's shopping lists. A list belongs to the account that created it but is visible to
// and editable by that account's co-parents (parents sharing a child) -- see internal/shopping. A finished list stays in
// the history ("completed"): it can be browsed, copied into a new list, and feeds the "often bought" suggestions.
package shoppingpg

import "time"

const (
	StatusActive    = "active"
	StatusCompleted = "completed"
)

type List struct {
	Id            string     `json:"id" gorm:"primaryKey"`
	AccountId     string     `json:"account_id" gorm:"not null;index"`
	Name          string     `json:"name" gorm:"not null"`
	Status        string     `json:"status" gorm:"not null;default:active;index"`
	CreatedByName string     `json:"created_by_name"`
	CreateTime    time.Time  `json:"create_time"`
	CompletedTime *time.Time `json:"completed_time,omitempty"`
}

func (List) TableName() string { return "shopping_lists" }

type Item struct {
	Id          string     `json:"id" gorm:"primaryKey"`
	ListId      string     `json:"list_id" gorm:"index"`
	AccountId   string     `json:"account_id" gorm:"not null;index"`
	Name        string     `json:"name" gorm:"not null"`
	Quantity    string     `json:"quantity"`
	Done        bool       `json:"done" gorm:"not null;default:false"`
	DoneBy      string     `json:"done_by,omitempty"`
	DoneAt      *time.Time `json:"done_at,omitempty"`
	AddedByName string     `json:"added_by_name"`
	CreateTime  time.Time  `json:"create_time"`
}

func (Item) TableName() string { return "shopping_items" }

// ListSummary is a list with its counters, for the overview and history screens.
type ListSummary struct {
	List
	Total int `json:"total"`
	Done  int `json:"done"`
}

// ListDetail is one list with its items plus suggestions ("often bought", minus what is already on the list).
type ListDetail struct {
	List        ListSummary `json:"list"`
	Items       []Item      `json:"items"`
	Suggestions []string    `json:"suggestions"`
}

// HistoryEntry is a product the family has bought/added before, most frequent first.
type HistoryEntry struct {
	Name         string `json:"name"`
	Count        int    `json:"count"`
	LastQuantity string `json:"last_quantity"`
}

type CreateListRequest struct {
	Name string `json:"name"`
}

type UpdateListRequest struct {
	Name   *string `json:"name,omitempty"`
	Status *string `json:"status,omitempty"`
}

type CopyListRequest struct {
	Name string `json:"name"`
}

type ItemInput struct {
	Name     string `json:"name"`
	Quantity string `json:"quantity"`
}

type AddItemsRequest struct {
	ListId string      `json:"list_id"`
	Items  []ItemInput `json:"items"`
}

type UpdateItemRequest struct {
	Done     *bool   `json:"done,omitempty"`
	Name     *string `json:"name,omitempty"`
	Quantity *string `json:"quantity,omitempty"`
}
