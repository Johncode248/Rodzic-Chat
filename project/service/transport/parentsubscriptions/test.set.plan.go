package parentsubscriptionsgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	parentsubscriptionspg "github.com/Johncode248/Rodzic-Chat/project/common/parentsubscriptions"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentsubscriptions"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
	"github.com/Johncode248/Rodzic-Chat/variables"
)

// TestSetPlan handles POST /private/parent-subscription/test-set-plan --
// lets the CALLER change their OWN plan directly, no payment at all.
//
// TEMPORARY, BY DESIGN: Apple/Google require in-app purchase (StoreKit /
// Play Billing) for this kind of digital entitlement -- self-granting a
// paid plan for free must never ship in a production build. This exists
// only to unblock building/testing the plan+limits UI before real IAP is
// wired up. Delete this endpoint (or gate it behind a build flag) before
// any App Store/Play Store submission.
func TestSetPlan(w http.ResponseWriter, r *http.Request) {
	var (
		request parentsubscriptionspg.TestSetPlanRequest
		res     parentsubscriptionspg.StatusResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "parentsubscriptionsgt.TestSetPlan", r.Body)
	defer logger.Exit(ctx, "parentsubscriptionsgt.TestSetPlan", &res, &err)

	if !variables.AllowTestEndpoints {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	plan := parentsubscriptionspg.MapPlan(request.Plan)
	subscription, err := parentsubscriptions.TestSetPlan(ctx, accountId, plan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res = parentsubscriptionspg.StatusResponse{
		Plan:            subscription.Plan.String(),
		AiRequestsUsed:  subscription.AiRequestsUsed,
		AiRequestsLimit: parentsubscriptionspg.MonthlyAiRequestLimit(subscription.Plan),
		PeriodResetAt:   subscription.PeriodResetAt.Format("2006-01-02"),
	}

	json.NewEncoder(w).Encode(res)
}
