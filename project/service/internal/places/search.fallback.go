package places

import (
	"context"

	placespg "github.com/Johncode248/Rodzic-Chat/project/common/places"
	"github.com/Johncode248/Rodzic-Chat/variables"
)

// SearchWithFallback is THE place search entry point every parent goes through now (Google
// Places is the primary source for everyone -- MVP has no plan gate on it, see
// variables.GooglePlacesApiKey's doc comment). It tries Google first, but only when the shared
// global monthly budget (placespg.MonthlyGooglePlacesRequestBudget) has room -- there is
// deliberately no per-parent daily cap any more (ParentSubscription was stripped down to just
// Plan for MVP; see its doc comment), so one very active parent can exhaust the shared budget
// alone. Budget exhaustion, Google not being configured, or the Google call itself erroring all
// fall through to the always-free OpenStreetMap search instead of failing the parent's search
// outright. [provider] reports which one actually served the results ("google" or "osm").
func SearchWithFallback(ctx context.Context, accountId string, query string) (candidates []placespg.SearchCandidate, provider string, err error) {
	if variables.GooglePlacesApiKey != "" {
		globalRoom, budgetErr := globalGoogleBudgetHasRoom(ctx)
		if budgetErr == nil && globalRoom {
			if googleCandidates, googleErr := SearchGoogle(ctx, query); googleErr == nil {
				_ = incrementGlobalGoogleUsage(ctx)
				return googleCandidates, placespg.SourceGoogle, nil
			}
		}
	}

	candidates, err = Search(ctx, query)
	return candidates, placespg.SourceOSM, err
}
