package parentchildgt

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentchild"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

var errNotYourParentChildLink = errors.New("you may only remove your own parent-child links")

// Delete requires the caller's Query to include a "child_id" filter so we can
// verify the child belongs to the caller's organization before deleting. A
// non-staff caller (a parent) may only delete links where they are the
// parent -- staff (director/co-director/teacher) may unlink any pair in
// their org.
func Delete(w http.ResponseWriter, r *http.Request) {
	var (
		request parentchildpg.DeleteRequest
		res     parentchildpg.DeleteResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "parentchildgt.Delete", r.Body)
	defer logger.Exit(ctx, "parentchildgt.Delete", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	// ParentChild has no organization_id column of its own, so unlike
	// every other org-scoped Delete in this codebase there's no query-level
	// AddOrganizationAuth safety net -- the org check below on each
	// child_id IS the entire tenant boundary for this endpoint. The query
	// can be a multi-OR-group filter (each group ANDed internally, groups
	// ORed together at the SQL level -- see gormdb.makeCondition), so EVERY
	// group must carry its own validated child_id: a group that omits one
	// entirely would otherwise let that group's other filters (e.g. a bare
	// parent_id) match and delete rows completely unconstrained by
	// organization, regardless of what child_id appeared in some other
	// group. This previously only collected the union of child_ids seen
	// anywhere and validated those, which a query like
	// {"groups":[{"child_id":"<a real child in this org>"},{"parent_id":"<arbitrary>"}]}
	// bypassed entirely.
	if len(request.Query.Groups) == 0 {
		http.Error(w, "Query must filter by child_id", http.StatusBadRequest)
		return
	}

	childIdSet := map[string]bool{}
	for _, group := range request.Query.Groups {
		groupChildId := ""
		for _, filter := range group.Filters {
			if filter.Key == "child_id" {
				if value, ok := filter.Value.(string); ok && value != "" {
					groupChildId = value
				}
			}
		}
		if groupChildId == "" {
			http.Error(w, "Every filter group must include a child_id", http.StatusBadRequest)
			return
		}
		childIdSet[groupChildId] = true
	}

	for childId := range childIdSet {
		inOrg, err := childInOrganization(ctx, childId, organizationId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !inOrg {
			http.Error(w, errChildNotInOrganization.Error(), http.StatusForbidden)
			return
		}
	}

	accountType, err := credentials.GetAccountTypeFromContext(r.Context())
	if err != nil {
		http.Error(w, "Error getting accountType from context", http.StatusForbidden)
		return
	}

	isStaff := accountType == accountspg.AccountTypeDirector.Ordinal() ||
		accountType == accountspg.AccountTypeCoDirector.Ordinal() ||
		accountType == accountspg.AccountTypeTeacher.Ordinal()

	if !isStaff {
		accountId, err := credentials.GetUserIdFromContext(r.Context())
		if err != nil || accountId == "" {
			http.Error(w, "Error getting accountId from context", http.StatusForbidden)
			return
		}

		matching, err := parentchild.Select(ctx, parentchildpg.SelectArgument{Query: request.Query})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, link := range matching.ParentChildren {
			if link.ParentId != accountId {
				http.Error(w, errNotYourParentChildLink.Error(), http.StatusForbidden)
				return
			}
		}
	}

	res, err = parentchild.Delete(ctx, parentchildpg.DeleteArgument{
		Query: request.Query,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
