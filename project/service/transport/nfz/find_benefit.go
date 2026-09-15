package nfzgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	nfzpg "github.com/Johncode248/Rodzic-Chat/project/common/nfz"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/nfz"
)

// FindBenefit handles GET /private/nfz/benefits?query= -- resolves a parent's free-text search
// (e.g. "alergolog dla dziecka") into exact NFZ benefit names, which SearchQueues then requires.
func FindBenefit(w http.ResponseWriter, r *http.Request) {
	var (
		res nfzpg.FindBenefitResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "nfzgt.FindBenefit", r.URL.RawQuery)
	defer logger.Exit(ctx, "nfzgt.FindBenefit", &res, &err)

	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "query is required", http.StatusBadRequest)
		return
	}

	benefits, err := nfz.FindBenefit(ctx, query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	res = nfzpg.FindBenefitResponse{Benefits: benefits}
	json.NewEncoder(w).Encode(res)
}
