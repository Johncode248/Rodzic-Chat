package groupsgt

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	groupspg "github.com/Johncode248/Rodzic-Chat/project/common/groups"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/groups"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

var errTeacherNotInOrganization = errors.New("teacher account does not belong to the caller's organization")

func Create(w http.ResponseWriter, r *http.Request) {
	var (
		request groupspg.CreateRequest
		res     groupspg.CreateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "groupsgt.Create", r.Body)
	defer logger.Exit(ctx, "groupsgt.Create", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	if request.TeacherId != "" {
		ok, teacherErr := teacherInOrganization(ctx, request.TeacherId, organizationId)
		if teacherErr != nil {
			http.Error(w, teacherErr.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, errTeacherNotInOrganization.Error(), http.StatusForbidden)
			return
		}
	}

	res, err = groups.Create(ctx, groupspg.CreateArgument{
		OrganizationId:                    organizationId,
		Name:                              request.Name,
		TeacherId:                         request.TeacherId,
		RequireChildBirthDateVerification: request.RequireChildBirthDateVerification,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
