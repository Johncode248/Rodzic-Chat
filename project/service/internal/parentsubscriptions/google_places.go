package parentsubscriptions

import (
	"context"
	"time"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	parentsubscriptionspg "github.com/Johncode248/Rodzic-Chat/project/common/parentsubscriptions"
)

func isSameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// TryConsumeDailyGoogleSearch is the per-parent half of the Google Places gate (see
// project/service/internal/places.SearchWithFallback for the other half, the shared global
// monthly budget) -- rolls GooglePlacesUsedToday back to 0 if the stored date isn't today, then
// reports whether accountId has room left today and, if so, consumes one. Returns (false, nil)
// -- not an error -- when the daily cap is simply reached; the caller falls back to OpenStreetMap
// either way.
func TryConsumeDailyGoogleSearch(ctx context.Context, accountId string) (allowed bool, err error) {
	ctx = logger.Enter(ctx, "parentsubscriptions.TryConsumeDailyGoogleSearch", accountId)
	defer logger.Exit(ctx, "parentsubscriptions.TryConsumeDailyGoogleSearch", &allowed, &err)

	current, err := GetOrCreate(ctx, accountId)
	if err != nil {
		return false, err
	}

	now := time.Now()
	if !isSameDay(current.GooglePlacesUsageDate, now) {
		current.GooglePlacesUsedToday = 0
		current.GooglePlacesUsageDate = now
	}

	if current.GooglePlacesUsedToday >= parentsubscriptionspg.DailyGooglePlacesSearchLimit {
		current.UpdateTime = now
		if err = gormdb.Update(ctx, &current, byAccountId(accountId)); err != nil {
			return false, err
		}
		return false, nil
	}

	current.GooglePlacesUsedToday++
	current.UpdateTime = now
	if err = gormdb.Update(ctx, &current, byAccountId(accountId)); err != nil {
		return false, err
	}
	return true, nil
}
