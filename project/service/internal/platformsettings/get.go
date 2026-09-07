package platformsettings

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	platformsettingspg "github.com/Johncode248/Rodzic-Chat/project/common/platformsettings"
)

// Get returns the singleton PlatformSettings row, creating it with
// Defaults() on first use so callers never have to special-case "not
// configured yet".
func Get(ctx context.Context) (res platformsettingspg.PlatformSettings, err error) {
	ctx = logger.Enter(ctx, "platformsettings.Get")
	defer logger.Exit(ctx, "platformsettings.Get", &res, &err)

	query := databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: platformsettingspg.SingletonId}}},
		},
	}

	rows, err := gormdb.Select[platformsettingspg.PlatformSettings](ctx, query)
	if err != nil {
		return res, err
	}

	if len(rows) > 0 {
		return rows[0], nil
	}

	res = platformsettingspg.Defaults()
	if err = gormdb.Create(ctx, &res); err != nil {
		return platformsettingspg.PlatformSettings{}, err
	}

	return res, nil
}

// Update applies a partial change to the singleton row (creating it with
// Defaults() first if needed) and returns the result.
func Update(ctx context.Context, request platformsettingspg.UpdateRequest) (res platformsettingspg.PlatformSettings, err error) {
	ctx = logger.Enter(ctx, "platformsettings.Update", request)
	defer logger.Exit(ctx, "platformsettings.Update", &res, &err)

	current, err := Get(ctx)
	if err != nil {
		return res, err
	}

	if request.PlatformFeeGrosze != nil {
		current.PlatformFeeGrosze = *request.PlatformFeeGrosze
	}
	if request.StripeAssumedPercentBps != nil {
		current.StripeAssumedPercentBps = *request.StripeAssumedPercentBps
	}
	if request.StripeAssumedFixedGrosze != nil {
		current.StripeAssumedFixedGrosze = *request.StripeAssumedFixedGrosze
	}

	query := databasepg.Query{
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: platformsettingspg.SingletonId}}},
		},
	}

	if err = gormdb.Update(ctx, &current, query); err != nil {
		return res, err
	}

	return current, nil
}
