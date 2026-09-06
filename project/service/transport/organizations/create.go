package organizationsgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/organizations"
)

func Create(w http.ResponseWriter, r *http.Request) {
	var (
		request organizationspg.CreateRequest
		res     organizationspg.CreateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "organizationsgt.Create", r.Body)
	defer logger.Exit(ctx, "organizationsgt.Create", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	organizationType, err := authorizationpg.MapOrganizationType(request.Type)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err = organizations.Create(ctx, organizationspg.CreateArgument{
		Name:         request.Name,
		Type:         organizationType,
		Subdomain:    request.Subdomain,
		Address:      request.Address,
		Nip:          request.Nip,
		LogoUrl:      request.LogoUrl,
		PrimaryColor: request.PrimaryColor,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
