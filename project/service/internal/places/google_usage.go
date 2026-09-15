package places

import (
	"context"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	placespg "github.com/Johncode248/Rodzic-Chat/project/common/places"
	"github.com/google/uuid"
)

func byPeriodMonth(month string) databasepg.Query {
	return databasepg.Query{Limit: 1, Groups: []databasepg.Group{
		{Filters: []databasepg.Filter{{Key: "period_month", Operator: "=", Value: month}}},
	}}
}

func currentPeriodMonth() string {
	return time.Now().Format("2006-01")
}

// getOrCreateGoogleUsage returns (creating if needed) this calendar month's global Google
// Places counter row -- see placespg.GooglePlacesUsage's doc comment.
func getOrCreateGoogleUsage(ctx context.Context) (res placespg.GooglePlacesUsage, err error) {
	month := currentPeriodMonth()
	existing, err := gormdb.Select[placespg.GooglePlacesUsage](ctx, byPeriodMonth(month))
	if err != nil {
		return res, err
	}
	if len(existing) > 0 {
		return existing[0], nil
	}

	now := time.Now()
	fresh := placespg.GooglePlacesUsage{
		Id:          uuid.NewString(),
		PeriodMonth: month,
		CreateTime:  now,
		UpdateTime:  now,
	}
	if err = gormdb.Create(ctx, &fresh); err != nil {
		return res, err
	}
	return fresh, nil
}

// globalGoogleBudgetHasRoom reports whether the shared monthly Google Places budget
// (placespg.MonthlyGooglePlacesRequestBudget) still has room this month.
func globalGoogleBudgetHasRoom(ctx context.Context) (bool, error) {
	usage, err := getOrCreateGoogleUsage(ctx)
	if err != nil {
		return false, err
	}
	return usage.RequestsUsed < placespg.MonthlyGooglePlacesRequestBudget, nil
}

// incrementGlobalGoogleUsage bumps this month's shared counter by one -- called only after a
// Google Places call actually succeeds, so a failed/erroring call never burns shared budget for
// nothing.
func incrementGlobalGoogleUsage(ctx context.Context) error {
	usage, err := getOrCreateGoogleUsage(ctx)
	if err != nil {
		return err
	}
	usage.RequestsUsed++
	usage.UpdateTime = time.Now()
	return gormdb.Update(ctx, &usage, byPeriodMonth(usage.PeriodMonth))
}
