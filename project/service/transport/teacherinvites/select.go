package teacherinvitesgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	teacherinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/teacherinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/teacherinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/pagination"
)

// Select lists every teacher invite in the caller's organization,
// director-only, optionally filtered by ?group_id=.
func Select(w http.ResponseWriter, r *http.Request) {
	var (
		res teacherinvitespg.SelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "teacherinvitesgt.Select")
	defer logger.Exit(ctx, "teacherinvitesgt.Select", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	query := databasepg.Query{
		Groups: []databasepg.Group{{}},
	}

	if groupId := r.URL.Query().Get("group_id"); groupId != "" {
		query.Groups[0].Filters = append(query.Groups[0].Filters, databasepg.Filter{
			Key: "group_id", Operator: "=", Value: groupId,
		})
	}

	query = query.AddOrganizationAuth(organizationId)
	query.OrderBy = databasepg.Order{Key: "create_time", Direction: databasepg.DirectionDescending}
	query.Limit, query.Offset = pagination.Parse(r, pagination.DefaultMaxLimit)

	res, err = teacherinvites.Select(ctx, teacherinvitespg.SelectArgument{Query: query})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
