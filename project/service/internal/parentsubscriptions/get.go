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

// GetOrCreate returns accountId's ParentSubscription, creating a fresh PlanFree row on first
// call.
func GetOrCreate(ctx context.Context, accountId string) (res parentsubscriptionspg.ParentSubscription, err error) {
	ctx = logger.Enter(ctx, "parentsubscriptions.GetOrCreate", accountId)
	defer logger.Exit(ctx, "parentsubscriptions.GetOrCreate", &res, &err)

	existing, err := gormdb.Select[parentsubscriptionspg.ParentSubscription](ctx, byAccountId(accountId))
	if err != nil {
		return res, err
	}

	if len(existing) > 0 {
		return existing[0], nil
	}

	now := time.Now()
	fresh := parentsubscriptionspg.ParentSubscription{
		Id:         uuid.NewString(),
		AccountId:  accountId,
		Plan:       parentsubscriptionspg.PlanFree,
		CreateTime: now,
		UpdateTime: now,
	}
	if err = gormdb.Create(ctx, &fresh); err != nil {
		return res, err
	}
	return fresh, nil
}
