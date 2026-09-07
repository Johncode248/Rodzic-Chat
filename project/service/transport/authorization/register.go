package authgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/authorization"
)

func Register(w http.ResponseWriter, r *http.Request) {
	var (
		request authorizationpg.RegisterRequest
		res     authorizationpg.RegisterResponse
		err     error
	)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	ctx := logger.Enter(r.Context(), "authgt.Register", &request)
	defer logger.Exit(ctx, "authgt.Register", &res, &err)

	result, err := authorization.Register(ctx, authorizationpg.RegisterArgument{
		Name:                      request.Name,
		Email:                     request.Email,
		Phone:                     request.Phone,
		Password:                  request.Password,
		OrganizationName:          request.OrganizationName,
		OrganizationType:          request.OrganizationType,
		OrganizationSubdomain:     request.OrganizationSubdomain,
		OrganizationChildrenLimit: request.OrganizationChildrenLimit,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res = result.Inherit()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
