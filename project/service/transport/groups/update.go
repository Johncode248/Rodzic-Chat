package groupsgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	groupspg "github.com/Johncode248/Rodzic-Chat/project/common/groups"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/groups"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

func Update(w http.ResponseWriter, r *http.Request) {
	var (
		request groupspg.UpdateRequest
		res     groupspg.UpdateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "groupsgt.Update", r.Body)
	defer logger.Exit(ctx, "groupsgt.Update", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	res, err = groups.Update(ctx, groupspg.UpdateArgument{
		Query:                             request.Query.AddOrganizationAuth(organizationId),
		Name:                              request.Name,
		TeacherId:                         request.TeacherId,
		AgeRange:                          request.AgeRange,
		RequireChildBirthDateVerification: request.RequireChildBirthDateVerification,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
