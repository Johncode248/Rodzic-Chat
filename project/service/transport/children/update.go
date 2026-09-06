package childrengt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

func Update(w http.ResponseWriter, r *http.Request) {
	var (
		request childrenpg.UpdateRequest
		res     childrenpg.UpdateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "childrengt.Update", r.Body)
	defer logger.Exit(ctx, "childrengt.Update", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	query := request.Query.AddOrganizationAuth(organizationId)

	res, err = children.Update(ctx, childrenpg.UpdateArgument{
		Query:     query,
		GroupId:   request.GroupId,
		FirstName: request.FirstName,
		LastName:  request.LastName,
		BirthDate: request.BirthDate,
		Notes:     request.Notes,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
