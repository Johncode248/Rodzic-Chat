package groupteachersgt

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	groupteacherspg "github.com/Johncode248/Rodzic-Chat/project/common/groupteachers"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
	"github.com/google/uuid"
)

var errAlreadyAssigned = errors.New("this teacher is already assigned to this group")

// Create adds a teacher to a group's member list, beyond that group's single
// "primary" TeacherId. Staff-only (gated in module.go).
func Create(w http.ResponseWriter, r *http.Request) {
	var (
		request groupteacherspg.CreateRequest
		res     groupteacherspg.CreateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "groupteachersgt.Create", r.Body)
	defer logger.Exit(ctx, "groupteachersgt.Create", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	if request.GroupId == "" || request.TeacherId == "" {
		http.Error(w, "group_id and teacher_id are required", http.StatusBadRequest)
		return
	}

	groupOk, err := groupInOrganization(ctx, request.GroupId, organizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !groupOk {
		http.Error(w, errGroupNotInOrganization.Error(), http.StatusForbidden)
		return
	}

	teacherOk, err := teacherInOrganization(ctx, request.TeacherId, organizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !teacherOk {
		http.Error(w, errTeacherNotInOrganization.Error(), http.StatusForbidden)
		return
	}

	existing, err := gormdb.Select[groupteacherspg.GroupTeacher](ctx, databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{
				Filters: []databasepg.Filter{
					{Key: "group_id", Operator: "=", Value: request.GroupId},
					{Key: "teacher_id", Operator: "=", Value: request.TeacherId},
				},
			},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(existing) > 0 {
		http.Error(w, errAlreadyAssigned.Error(), http.StatusConflict)
		return
	}

	groupTeacher := groupteacherspg.GroupTeacher{
		Id:         uuid.NewString(),
		GroupId:    request.GroupId,
		TeacherId:  request.TeacherId,
		CreateTime: time.Now(),
	}

	if err = gormdb.Create(ctx, &groupTeacher); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res = groupteacherspg.CreateResponse{GroupTeacher: groupTeacher}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
