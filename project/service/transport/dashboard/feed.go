package dashboardgt

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Johncode248/Rodzic-Chat/logger"
	dashboardpg "github.com/Johncode248/Rodzic-Chat/project/common/dashboard"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/dashboard"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Feed returns the director dashboard's activity feed, org-scoped,
// merging recent new-parent registrations, absence reports, and chat
// announcements sorted by time desc. Optional ?limit= (default 20).
func Feed(w http.ResponseWriter, r *http.Request) {
	var (
		res dashboardpg.FeedResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "dashboardgt.Feed")
	defer logger.Exit(ctx, "dashboardgt.Feed", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	limit := 20
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 {
		limit = l
	}

	res, err = dashboard.Feed(ctx, organizationId, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
