package parentchildgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentchild"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Select lists parent-child links scoped to the caller's organization
// (via the child's organization_id, since parent_child has no organization_id
// column of its own), optionally narrowed by ?parent_id= or ?child_id=.
func Select(w http.ResponseWriter, r *http.Request) {
	var (
		res parentchildpg.SelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "parentchildgt.Select")
	defer logger.Exit(ctx, "parentchildgt.Select", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	filters := []databasepg.Filter{}

	if childId := r.URL.Query().Get("child_id"); childId != "" {
		inOrg, err := childInOrganization(ctx, childId, organizationId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !inOrg {
			http.Error(w, errChildNotInOrganization.Error(), http.StatusForbidden)
			return
		}

		filters = append(filters, databasepg.Filter{Key: "child_id", Operator: "=", Value: childId})
	} else {
		childIds, err := childIdsInOrganization(ctx, organizationId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		filters = append(filters, databasepg.Filter{Key: "child_id", Operator: "IN", Value: childIds})
	}

	if parentId := r.URL.Query().Get("parent_id"); parentId != "" {
		filters = append(filters, databasepg.Filter{Key: "parent_id", Operator: "=", Value: parentId})
	}

	res, err = parentchild.Select(ctx, parentchildpg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: filters},
			},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
