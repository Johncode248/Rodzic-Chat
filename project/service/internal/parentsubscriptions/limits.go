package parentsubscriptions

import (
	"context"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	parentsubscriptionspg "github.com/Johncode248/Rodzic-Chat/project/common/parentsubscriptions"
	placespg "github.com/Johncode248/Rodzic-Chat/project/common/places"
)

func byAccountIdRecommendations(accountId string) databasepg.Query {
	return databasepg.Query{Groups: []databasepg.Group{
		{Filters: []databasepg.Filter{{Key: "account_id", Operator: "=", Value: accountId}}},
	}}
}

// Limits gathers accountId's "Limity" screen data in one call -- AI tokens, recommendations
// made, and today's Google Places searches, each against its own cap. Rolls the monthly/daily
// counters forward first (via GetOrCreate) so the numbers are always current, never stale from
// a previous period.
func Limits(ctx context.Context, accountId string) (res parentsubscriptionspg.LimitsResponse, err error) {
	ctx = logger.Enter(ctx, "parentsubscriptions.Limits", accountId)
	defer logger.Exit(ctx, "parentsubscriptions.Limits", &res, &err)

	subscription, err := GetOrCreate(ctx, accountId)
	if err != nil {
		return res, err
	}

	recommendations, err := gormdb.Select[placespg.Recommendation](ctx, byAccountIdRecommendations(accountId))
	if err != nil {
		return res, err
	}

	googlePlacesToday := subscription.GooglePlacesUsedToday
	if !isSameDay(subscription.GooglePlacesUsageDate, time.Now()) {
		// The stored counter is from a previous day -- report 0 rather than a stale number; the
		// next actual search rolls it forward for real (see TryConsumeDailyGoogleSearch).
		googlePlacesToday = 0
	}

	return parentsubscriptionspg.LimitsResponse{
		Plan:                      subscription.Plan.String(),
		AiTokensUsed:              subscription.AiTokensUsed,
		AiTokensLimit:             parentsubscriptionspg.MonthlyAiTokenLimit(subscription.Plan),
		PeriodResetAt:             subscription.PeriodResetAt.Format("2006-01-02"),
		RecommendationsUsed:       int64(len(recommendations)),
		RecommendationsLimit:      parentsubscriptionspg.MaxRecommendationsPerParent,
		GooglePlacesSearchesToday: googlePlacesToday,
		GooglePlacesDailyLimit:    parentsubscriptionspg.DailyGooglePlacesSearchLimit,
	}, nil
}
