package parentchildgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentchild"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

func Create(w http.ResponseWriter, r *http.Request) {
	var (
		request parentchildpg.CreateRequest
		res     parentchildpg.CreateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "parentchildgt.Create", r.Body)
	defer logger.Exit(ctx, "parentchildgt.Create", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	inOrg, err := childInOrganization(ctx, request.ChildId, organizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !inOrg {
		http.Error(w, errChildNotInOrganization.Error(), http.StatusForbidden)
		return
	}

	parentOk, err := parentInOrganization(ctx, request.ParentId, organizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !parentOk {
		http.Error(w, errParentNotInOrganization.Error(), http.StatusForbidden)
		return
	}

	res, err = parentchild.Create(ctx, parentchildpg.CreateArgument{
		ParentId:  request.ParentId,
		ChildId:   request.ChildId,
		Relation:  parentchildpg.MapRelationType(request.Relation),
		CanPickup: request.CanPickup,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
