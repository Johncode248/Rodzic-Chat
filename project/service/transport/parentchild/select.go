package parentchildgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentchild"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/pagination"
)

// Select lists parent-child links scoped to the caller's organization
// (via the child's organization_id, since parent_child has no organization_id
// column of its own), optionally narrowed by ?parent_id= or ?child_id=.
// Staff (director/co-director/teacher) see every link in the org; a parent
// only ever sees links for children they themselves are linked to -- this
// endpoint effectively exposes who else has pickup rights over a child, so
// it must never be usable to enumerate every family's arrangements in the
// placówka.
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

	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	accountType, err := credentials.GetAccountTypeFromContext(r.Context())
	if err != nil {
		http.Error(w, "Error getting accountType from context", http.StatusForbidden)
		return
	}
	isStaff := accountType == accountspg.AccountTypeDirector.Ordinal() || accountType == accountspg.AccountTypeCoDirector.Ordinal() || accountType == accountspg.AccountTypeTeacher.Ordinal()

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

		if !isStaff {
			owns, err := parentOwnsChild(ctx, accountId, childId)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if !owns {
				http.Error(w, errNotChildsParent.Error(), http.StatusForbidden)
				return
			}
		}

		filters = append(filters, databasepg.Filter{Key: "child_id", Operator: "=", Value: childId})
	} else if isStaff {
		childIds, err := childIdsInOrganization(ctx, organizationId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		filters = append(filters, databasepg.Filter{Key: "child_id", Operator: "IN", Value: childIds})
	} else {
		childIds, err := childIdsForParent(ctx, accountId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		filters = append(filters, databasepg.Filter{Key: "child_id", Operator: "IN", Value: childIds})
	}

	if parentId := r.URL.Query().Get("parent_id"); parentId != "" {
		filters = append(filters, databasepg.Filter{Key: "parent_id", Operator: "=", Value: parentId})
	}

	pcQuery := databasepg.Query{
		Groups:  []databasepg.Group{{Filters: filters}},
		OrderBy: databasepg.Order{Key: "create_time", Direction: databasepg.DirectionDescending},
	}
	pcQuery.Limit, pcQuery.Offset = pagination.Parse(r, pagination.DefaultMaxLimit)

	res, err = parentchild.Select(ctx, parentchildpg.SelectArgument{Query: pcQuery})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
