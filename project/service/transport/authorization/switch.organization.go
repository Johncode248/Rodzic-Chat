package authgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/authorization"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

func SwitchOrganization(w http.ResponseWriter, r *http.Request) {
	var (
		request authorizationpg.SwitchOrganizationRequest
		res     authorizationpg.SwitchOrganizationResponse
		err     error
	)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	ctx := logger.Enter(r.Context(), "authgt.SwitchOrganization", &request)
	defer logger.Exit(ctx, "authgt.SwitchOrganization", &res, &err)

	userId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil {
		http.Error(w, "Auth went wrong", http.StatusForbidden)
		return
	}

	email, err := credentials.GetUserEmailFromContext(ctx)
	if err != nil {
		http.Error(w, "Auth went wrong", http.StatusForbidden)
		return
	}

	sessionId, err := credentials.GetSessionIdFromContext(ctx)
	if err != nil {
		http.Error(w, "Auth went wrong", http.StatusForbidden)
		return
	}

	res, err = authorization.SwitchOrganization(ctx, authorizationpg.SwitchOrganizationArgument{
		AccountId:      userId,
		Email:          email,
		SessionId:      sessionId,
		OrganizationId: request.OrganizationId,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(res)
}
