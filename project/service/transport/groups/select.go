package groupsgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	groupspg "github.com/Johncode248/Rodzic-Chat/project/common/groups"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/groups"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/pagination"
)

// Select lists groups in the caller's organization, optionally filtered by
// ?q= (free-text, group name, case-insensitive substring match).
func Select(w http.ResponseWriter, r *http.Request) {
	var (
		res groupspg.SelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "groupsgt.Select")
	defer logger.Exit(ctx, "groupsgt.Select", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	query := databasepg.Query{
		Groups: []databasepg.Group{{}},
	}

	if q := r.URL.Query().Get("q"); q != "" {
		query.Groups[0].Filters = append(query.Groups[0].Filters, databasepg.Filter{
			Key: "name", Operator: "ILIKE", Value: "%" + q + "%",
		})
	}

	query = query.AddOrganizationAuth(organizationId)
	query.OrderBy = databasepg.Order{Key: "name", Direction: databasepg.DirectionAscending}
	query.Limit, query.Offset = pagination.Parse(r, pagination.DefaultMaxLimit)

	res, err = groups.Select(ctx, groupspg.SelectArgument{
		Query: query,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
