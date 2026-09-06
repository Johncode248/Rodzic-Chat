package childrengt

import (
	"encoding/json"
	"net/http"

	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"

	"github.com/Johncode248/Rodzic-Chat/logger"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

func Delete(w http.ResponseWriter, r *http.Request) {
	var (
		request childrenpg.DeleteRequest
		res     childrenpg.DeleteResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "childrengt.Delete", r.Body)
	defer logger.Exit(ctx, "childrengt.Delete", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	res, err = children.Delete(ctx, childrenpg.DeleteArgument{
		Query: request.Query.AddOrganizationAuth(organizationId),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
