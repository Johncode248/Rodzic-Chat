package parentsubscriptions

import (
	"context"

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

// Limits gathers accountId's "Limity" screen data in one call. AI usage and today's Google
// Places searches always report 0 used now -- ParentSubscription no longer tracks either (the AI
// assistant endpoint is hard-disabled server-side, and Google Places is only gated by the shared
// global monthly budget now, not a per-parent daily one -- see their doc comments in
// project/common/parentsubscriptions). RecommendationsUsed is unaffected: it's counted directly
// from project/common/places.Recommendation, never stored on ParentSubscription at all.
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

	return parentsubscriptionspg.LimitsResponse{
		Plan:                      subscription.Plan.String(),
		AiTokensUsed:              0,
		AiTokensLimit:             parentsubscriptionspg.MonthlyAiTokenLimit(subscription.Plan),
		RecommendationsUsed:       int64(len(recommendations)),
		RecommendationsLimit:      parentsubscriptionspg.MaxRecommendationsPerParent,
		GooglePlacesSearchesToday: 0,
		GooglePlacesDailyLimit:    parentsubscriptionspg.DailyGooglePlacesSearchLimit,
	}, nil
}
