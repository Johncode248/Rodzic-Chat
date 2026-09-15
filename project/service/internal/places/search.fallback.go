package places

import (
	"context"

	placespg "github.com/Johncode248/Rodzic-Chat/project/common/places"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentsubscriptions"
	"github.com/Johncode248/Rodzic-Chat/variables"
)

// SearchWithFallback is THE place search entry point every parent goes through now (Google
// Places is the primary source for everyone -- MVP has no plan gate on it, see
// variables.GooglePlacesApiKey's doc comment). It tries Google first, but only when BOTH the
// shared global monthly budget (placespg.MonthlyGooglePlacesRequestBudget) and this parent's own
// daily cap (parentsubscriptionspg.DailyGooglePlacesSearchLimit) have room; any of that being
// exhausted, Google not being configured, or the Google call itself erroring all fall through to
// the always-free OpenStreetMap search instead of failing the parent's search outright.
// [provider] reports which one actually served the results ("google" or "osm").
func SearchWithFallback(ctx context.Context, accountId string, query string) (candidates []placespg.SearchCandidate, provider string, err error) {
	if variables.GooglePlacesApiKey != "" {
		globalRoom, budgetErr := globalGoogleBudgetHasRoom(ctx)
		if budgetErr == nil && globalRoom {
			dailyAllowed, dailyErr := parentsubscriptions.TryConsumeDailyGoogleSearch(ctx, accountId)
			if dailyErr == nil && dailyAllowed {
				if googleCandidates, googleErr := SearchGoogle(ctx, query); googleErr == nil {
					_ = incrementGlobalGoogleUsage(ctx)
					return googleCandidates, placespg.SourceGoogle, nil
				}
			}
		}
	}

	candidates, err = Search(ctx, query)
	return candidates, placespg.SourceOSM, err
}
