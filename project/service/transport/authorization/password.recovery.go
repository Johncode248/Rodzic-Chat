package authgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/authorization"
)

func PasswordRecovery(w http.ResponseWriter, r *http.Request) {
	var (
		request authorizationpg.PasswordRecoveryRequest
		res     authorizationpg.PasswordRecoveryResponse
		err     error
	)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	ctx := logger.Enter(r.Context(), "authgt.PasswordRecovery", &request)
	defer logger.Exit(ctx, "authgt.PasswordRecovery", &res, &err)

	if err := authorization.SendPasswordResetEmail(ctx, request.Email); err != nil {
		http.Error(w, "Failed to send email", http.StatusConflict)
	}

	json.NewEncoder(w).Encode(res)
}
