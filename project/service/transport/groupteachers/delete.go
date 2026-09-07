package groupteachersgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	groupteacherspg "github.com/Johncode248/Rodzic-Chat/project/common/groupteachers"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Delete removes a teacher from a group's member list via ?group_id= and
// ?teacher_id= query params (both required). Staff-only (gated in
// module.go). Does not touch the group's separate "primary" TeacherId.
func Delete(w http.ResponseWriter, r *http.Request) {
	var (
		res groupteacherspg.DeleteResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "groupteachersgt.Delete")
	defer logger.Exit(ctx, "groupteachersgt.Delete", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	groupId := r.URL.Query().Get("group_id")
	teacherId := r.URL.Query().Get("teacher_id")
	if groupId == "" || teacherId == "" {
		http.Error(w, "group_id and teacher_id query params are required", http.StatusBadRequest)
		return
	}

	groupOk, err := groupInOrganization(ctx, groupId, organizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !groupOk {
		http.Error(w, errGroupNotInOrganization.Error(), http.StatusForbidden)
		return
	}

	result, err := gormdb.Delete[groupteacherspg.GroupTeacher](ctx, databasepg.Query{
		Groups: []databasepg.Group{
			{
				Filters: []databasepg.Filter{
					{Key: "group_id", Operator: "=", Value: groupId},
					{Key: "teacher_id", Operator: "=", Value: teacherId},
				},
			},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res = groupteacherspg.DeleteResponse{GroupTeachers: result}

	json.NewEncoder(w).Encode(res)
}
