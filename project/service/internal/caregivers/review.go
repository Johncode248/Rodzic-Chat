package caregivers

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	caregiverspg "github.com/Johncode248/Rodzic-Chat/project/common/caregivers"
)

// PendingItem is one row of the admin verification queue: the profile plus who owns it.
type PendingItem struct {
	caregiverspg.Caregiver
	OwnerName  string `json:"owner_name"`
	OwnerEmail string `json:"owner_email"`
	OwnerPhone string `json:"owner_phone"`
}

var ErrReasonRequired = errors.New("a reason is required when rejecting")

// ListForReview lists profiles by review status (1 pending, 2 approved, 3 rejected), oldest first for pending.
func ListForReview(ctx context.Context, status int) ([]PendingItem, error) {
	type row struct {
		caregiverspg.Caregiver
		OwnerName  string
		OwnerEmail string
		OwnerPhone string
	}
	rows, err := gormdb.CustomSelect[row](ctx, `select c.*, a.name as owner_name, a.email as owner_email, a.phone as owner_phone
 from caregivers c join accounts a on a.id = c.account_id
 where c.review_status = ? order by c.update_time asc`, status)
	if err != nil {
		return nil, err
	}
	out := make([]PendingItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, PendingItem{Caregiver: r.Caregiver, OwnerName: r.OwnerName, OwnerEmail: r.OwnerEmail, OwnerPhone: r.OwnerPhone})
	}
	return out, nil
}

// Decide approves or rejects a profile (a reason is mandatory for a rejection -- the owner sees it).
func Decide(ctx context.Context, id string, approve bool, reason string) error {
	c, err := getById(ctx, id)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrNotFound
	}
	now := time.Now()
	c.ReviewedAt = &now
	c.UpdateTime = now
	if approve {
		c.ReviewStatus = caregiverspg.ReviewStatusApproved
		c.RejectReason = ""
	} else {
		if reason == "" {
			return ErrReasonRequired
		}
		c.ReviewStatus = caregiverspg.ReviewStatusRejected
		c.RejectReason = reason
	}
	return gormdb.Update(ctx, *c, databasepg.Query{Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: id}}}}})
}
