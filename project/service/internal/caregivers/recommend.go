package caregivers

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	caregiverspg "github.com/Johncode248/Rodzic-Chat/project/common/caregivers"
	"github.com/google/uuid"
)

var ErrCannotRecommendSelf = errors.New("you cannot recommend your own profile")

// Recommend adds or updates the caller's "Polecam" endorsement of
// caregiverId -- recommending again just updates the comment (unique per
// caregiver+account, see CaregiverRecommendation's doc comment).
func Recommend(ctx context.Context, caregiverId, accountId, comment string) (res caregiverspg.CaregiverRecommendation, err error) {
	caregiver, err := getById(ctx, caregiverId)
	if err != nil {
		return res, err
	}
	if caregiver == nil {
		return res, ErrNotFound
	}
	if caregiver.AccountId == accountId {
		return res, ErrCannotRecommendSelf
	}
	if caregiver.ReviewStatus == caregiverspg.ReviewStatusPending || caregiver.ReviewStatus == caregiverspg.ReviewStatusRejected {
		return res, ErrNotFound
	}

	existing, err := gormdb.Select[caregiverspg.CaregiverRecommendation](ctx, databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{
				{Key: "caregiver_id", Operator: "=", Value: caregiverId},
				{Key: "account_id", Operator: "=", Value: accountId},
			}},
		},
	})
	if err != nil {
		return res, err
	}

	if len(existing) > 0 {
		updated := existing[0]
		updated.Comment = comment
		if err = gormdb.Update(ctx, updated, databasepg.Query{
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: updated.Id}}}},
		}); err != nil {
			return res, err
		}
		return updated, nil
	}

	created := caregiverspg.CaregiverRecommendation{
		Id:          uuid.NewString(),
		CaregiverId: caregiverId,
		AccountId:   accountId,
		Comment:     comment,
		CreateTime:  time.Now(),
	}
	if err = gormdb.Create(ctx, &created); err != nil {
		return res, err
	}
	return created, nil
}

// Unrecommend retracts the caller's own endorsement, if any (idempotent).
func Unrecommend(ctx context.Context, caregiverId, accountId string) error {
	_, err := gormdb.Delete[caregiverspg.CaregiverRecommendation](ctx, databasepg.Query{
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{
				{Key: "caregiver_id", Operator: "=", Value: caregiverId},
				{Key: "account_id", Operator: "=", Value: accountId},
			}},
		},
	})
	return err
}
