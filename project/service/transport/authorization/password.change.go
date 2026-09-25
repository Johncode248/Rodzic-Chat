package authgt

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/authorization"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// ChangePassword is POST /private/accounts/password -- authenticated self-service password change
// (current password required). Wrong current password -> 403 so clients can show a precise message.
func ChangePassword(w http.ResponseWriter, r *http.Request) {
	var (
		request authorizationpg.ChangePasswordRequest
		res     authorizationpg.ChangePasswordResponse
		err     error
	)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	ctx := logger.Enter(r.Context(), "authgt.ChangePassword", nil)
	defer logger.Exit(ctx, "authgt.ChangePassword", &res, &err)

	userId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || userId == "" {
		http.Error(w, "Error getting userId from context", http.StatusForbidden)
		return
	}
	sessionId, _ := credentials.GetSessionIdFromContext(r.Context())

	err = authorization.ChangePassword(ctx, authorizationpg.ChangePasswordArgument{
		AccountId:       userId,
		SessionId:       sessionId,
		CurrentPassword: request.CurrentPassword,
		NewPassword:     request.NewPassword,
	})
	switch {
	case errors.Is(err, authorization.ErrWrongCurrentPassword):
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(res)
}
