package parentsubscriptionsgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	parentsubscriptionspg "github.com/Johncode248/Rodzic-Chat/project/common/parentsubscriptions"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentsubscriptions"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Limits handles GET /private/parent-subscription/limits -- backs the "Limity" tab on Moje
// konto: AI tokens, recommendations made, and today's Google Places searches, each against its
// own cap (see parentsubscriptions.Limits).
func Limits(w http.ResponseWriter, r *http.Request) {
	var (
		res parentsubscriptionspg.LimitsResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "parentsubscriptionsgt.Limits")
	defer logger.Exit(ctx, "parentsubscriptionsgt.Limits", &res, &err)

	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	res, err = parentsubscriptions.Limits(ctx, accountId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
