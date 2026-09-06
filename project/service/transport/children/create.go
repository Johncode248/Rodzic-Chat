package childrengt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

func Create(w http.ResponseWriter, r *http.Request) {
	var (
		request childrenpg.CreateRequest
		res     childrenpg.CreateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "childrengt.Create", r.Body)
	defer logger.Exit(ctx, "childrengt.Create", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	res, err = children.Create(ctx, childrenpg.CreateArgument{
		OrganizationId: organizationId,
		GroupId:        request.GroupId,
		FirstName:      request.FirstName,
		LastName:       request.LastName,
		BirthDate:      request.BirthDate,
		Notes:          request.Notes,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
