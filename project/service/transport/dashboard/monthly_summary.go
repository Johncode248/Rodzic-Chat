package dashboardgt

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Johncode248/Rodzic-Chat/logger"
	dashboardpg "github.com/Johncode248/Rodzic-Chat/project/common/dashboard"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/dashboard"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// MonthlySummary handles GET /private/dashboard/monthly-summary -- a parent's own "Panel główny"
// digest (see dashboardpg.MonthlySummaryResponse's doc comment) for one calendar month. Optional
// ?year=&month= (both 1-indexed, month: 1=styczeń) default to the current month so the initial
// dashboard load needs no query params at all; the client passes them explicitly when the parent
// swipes to a different month.
func MonthlySummary(w http.ResponseWriter, r *http.Request) {
	var (
		res dashboardpg.MonthlySummaryResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "dashboardgt.MonthlySummary", r.URL.RawQuery)
	defer logger.Exit(ctx, "dashboardgt.MonthlySummary", &res, &err)

	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	now := time.Now()
	year := now.Year()
	month := int(now.Month())
	if y, convErr := strconv.Atoi(r.URL.Query().Get("year")); convErr == nil && y > 0 {
		year = y
	}
	if m, convErr := strconv.Atoi(r.URL.Query().Get("month")); convErr == nil && m >= 1 && m <= 12 {
		month = m
	}

	res, err = dashboard.MonthlySummary(ctx, accountId, year, month)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
