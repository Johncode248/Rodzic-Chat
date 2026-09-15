package parentsubscriptionsgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	parentsubscriptionspg "github.com/Johncode248/Rodzic-Chat/project/common/parentsubscriptions"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentsubscriptions"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Status handles GET /private/parent-subscription -- any authenticated
// account reads their own AI plan + current usage. Auto-creates a
// PlanFree row on first call (see parentsubscriptions.GetOrCreate).
func Status(w http.ResponseWriter, r *http.Request) {
	var (
		res parentsubscriptionspg.StatusResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "parentsubscriptionsgt.Status")
	defer logger.Exit(ctx, "parentsubscriptionsgt.Status", &res, &err)

	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	subscription, err := parentsubscriptions.GetOrCreate(ctx, accountId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res = parentsubscriptionspg.StatusResponse{
		Plan:            subscription.Plan.String(),
		AiRequestsUsed:  subscription.AiRequestsUsed,
		AiRequestsLimit: parentsubscriptionspg.MonthlyAiRequestLimit(subscription.Plan),
		AiTokensUsed:    subscription.AiTokensUsed,
		AiTokensLimit:   parentsubscriptionspg.MonthlyAiTokenLimit(subscription.Plan),
		PeriodResetAt:   subscription.PeriodResetAt.Format("2006-01-02"),
	}

	json.NewEncoder(w).Encode(res)
}
