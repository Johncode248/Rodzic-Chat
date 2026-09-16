package childrengt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/pagination"
)

// Select lists children in the caller's organization, optionally filtered by
// ?group_id= and/or ?q= (free-text, first/last name, case-insensitive
// substring match). Always scoped to the caller's organizationId, and --
// for a parent (non-staff) caller -- further scoped to only the children
// linked to them via ParentChild: staff need the full roster, but a parent
// must never be able to enumerate every other family's children in the
// same placówka.
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

	var ownChildIdFilter *databasepg.Filter
	isStaff := accountType == accountspg.AccountTypeDirector.Ordinal() || accountType == accountspg.AccountTypeCoDirector.Ordinal() || accountType == accountspg.AccountTypeTeacher.Ordinal()
	if !isStaff {
		childIds, err := childIdsForParent(ctx, accountId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ownChildIdFilter = &databasepg.Filter{Key: "id", Operator: "IN", Value: childIds}
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

	if ownChildIdFilter != nil {
		for i := range query.Groups {
			query.Groups[i].Filters = append(query.Groups[i].Filters, *ownChildIdFilter)
		}
	}

	query = query.AddOrganizationAuth(organizationId)
	query.OrderBy = databasepg.Order{Key: "first_name", Direction: databasepg.DirectionAscending}
	query.Limit, query.Offset = pagination.Parse(r, pagination.DefaultMaxLimit)

	res, err = children.Select(ctx, childrenpg.SelectArgument{
		Query: query,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
