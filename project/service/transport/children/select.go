package childrengt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Select lists children in the caller's organization, optionally filtered by
// ?group_id= and/or ?q= (free-text, first/last name, case-insensitive
// substring match). Always scoped to the caller's organizationId.
func Select(w http.ResponseWriter, r *http.Request) {
	var (
		res childrenpg.SelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "childrengt.Select")
	defer logger.Exit(ctx, "childrengt.Select", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	var groupFilter *databasepg.Filter
	if groupId := r.URL.Query().Get("group_id"); groupId != "" {
		groupFilter = &databasepg.Filter{Key: "group_id", Operator: "=", Value: groupId}
	}

	var query databasepg.Query

	if q := r.URL.Query().Get("q"); q != "" {
		like := "%" + q + "%"
		// One OR-group per searched column (first/last name) -- each gets
		// the group_id filter (if any) ANDed in below via
		// AddOrganizationAuth's existing "append to every group" behavior.
		query.Groups = []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "first_name", Operator: "ILIKE", Value: like}}},
			{Filters: []databasepg.Filter{{Key: "last_name", Operator: "ILIKE", Value: like}}},
		}
		if groupFilter != nil {
			for i := range query.Groups {
				query.Groups[i].Filters = append(query.Groups[i].Filters, *groupFilter)
			}
		}
	} else {
		query.Groups = []databasepg.Group{{}}
		if groupFilter != nil {
			query.Groups[0].Filters = append(query.Groups[0].Filters, *groupFilter)
		}
	}

	query = query.AddOrganizationAuth(organizationId)

	res, err = children.Select(ctx, childrenpg.SelectArgument{
		Query: query,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
