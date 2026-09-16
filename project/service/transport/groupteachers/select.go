package groupteachersgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	groupteacherspg "github.com/Johncode248/Rodzic-Chat/project/common/groupteachers"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/pagination"
)

// Select lists group-teacher memberships, optionally filtered by ?group_id=
// or ?teacher_id=. Always scoped to the caller's organization via the
// group side of the pair.
func Select(w http.ResponseWriter, r *http.Request) {
	var (
		res groupteacherspg.SelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "groupteachersgt.Select")
	defer logger.Exit(ctx, "groupteachersgt.Select", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	filters := []databasepg.Filter{}

	if groupId := r.URL.Query().Get("group_id"); groupId != "" {
		groupOk, err := groupInOrganization(ctx, groupId, organizationId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !groupOk {
			http.Error(w, errGroupNotInOrganization.Error(), http.StatusForbidden)
			return
		}
		filters = append(filters, databasepg.Filter{Key: "group_id", Operator: "=", Value: groupId})
	} else {
		groupIds, err := groupIdsInOrganization(ctx, organizationId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// A brand-new organization with zero groups yet would otherwise
		// build a "group_id IN ()" filter -- an empty IN list's behavior
		// isn't guaranteed safe/portable across query builders, so return
		// the (correctly) empty result directly instead of relying on it.
		if len(groupIds) == 0 {
			json.NewEncoder(w).Encode(groupteacherspg.SelectResponse{})
			return
		}
		filters = append(filters, databasepg.Filter{Key: "group_id", Operator: "IN", Value: groupIds})
	}

	if teacherId := r.URL.Query().Get("teacher_id"); teacherId != "" {
		filters = append(filters, databasepg.Filter{Key: "teacher_id", Operator: "=", Value: teacherId})
	}

	gtQuery := databasepg.Query{Groups: []databasepg.Group{{Filters: filters}}}
	gtQuery.Limit, gtQuery.Offset = pagination.Parse(r, pagination.DefaultMaxLimit)

	result, err := gormdb.Select[groupteacherspg.GroupTeacher](ctx, gtQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res = groupteacherspg.SelectResponse{GroupTeachers: result}

	json.NewEncoder(w).Encode(res)
}
