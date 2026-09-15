package parentsubscriptions

import (
	"context"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	parentsubscriptionspg "github.com/Johncode248/Rodzic-Chat/project/common/parentsubscriptions"
	"github.com/google/uuid"
)

func byAccountId(accountId string) databasepg.Query {
	return databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "account_id", Operator: "=", Value: accountId}}},
		},
	}
}

// nextPeriodReset is always "start of next calendar month" -- simple,
// predictable, matches how the placowka-side platform subscription already
// thinks about billing periods elsewhere in this codebase.
func nextPeriodReset(from time.Time) time.Time {
	firstOfThisMonth := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, from.Location())
	return firstOfThisMonth.AddDate(0, 1, 0)
}

// GetOrCreate returns accountId's ParentSubscription, creating a fresh
// PlanFree row on first call. Also resets AiRequestsUsed to 0 and rolls
// PeriodResetAt forward if the current period has elapsed -- callers never
// need to think about period rollover themselves.
func GetOrCreate(ctx context.Context, accountId string) (res parentsubscriptionspg.ParentSubscription, err error) {
	ctx = logger.Enter(ctx, "parentsubscriptions.GetOrCreate", accountId)
	defer logger.Exit(ctx, "parentsubscriptions.GetOrCreate", &res, &err)

	existing, err := gormdb.Select[parentsubscriptionspg.ParentSubscription](ctx, byAccountId(accountId))
	if err != nil {
		return res, err
	}

	now := time.Now()

	if len(existing) == 0 {
		fresh := parentsubscriptionspg.ParentSubscription{
			Id:             uuid.NewString(),
			AccountId:      accountId,
			Plan:           parentsubscriptionspg.PlanFree,
			AiRequestsUsed: 0,
			PeriodResetAt:  nextPeriodReset(now),
			CreateTime:     now,
			UpdateTime:     now,
		}
		if err = gormdb.Create(ctx, &fresh); err != nil {
			return res, err
		}
		return fresh, nil
	}

	subscription := existing[0]

	if !now.Before(subscription.PeriodResetAt) {
		subscription.AiRequestsUsed = 0
		subscription.PeriodResetAt = nextPeriodReset(now)
		subscription.UpdateTime = now
		if err = gormdb.Update(ctx, &subscription, byAccountId(accountId)); err != nil {
			return res, err
		}
	}

	return subscription, nil
}
