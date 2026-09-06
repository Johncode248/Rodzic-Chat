package authgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/authorization"
)

// ConsumePasswordReset is public/unauthenticated by design -- the token
// itself, emailed by PasswordRecovery, is the only credential needed (same
// reasoning already used for the registration-links public endpoints).
func ConsumePasswordReset(w http.ResponseWriter, r *http.Request) {
	var (
		request authorizationpg.ConsumePasswordResetRequest
		res     authorizationpg.ConsumePasswordResetResponse
		err     error
	)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	ctx := logger.Enter(r.Context(), "authgt.ConsumePasswordReset", &request)
	defer logger.Exit(ctx, "authgt.ConsumePasswordReset", &res, &err)

	err = authorization.ConsumePasswordReset(ctx, authorizationpg.ConsumePasswordResetArgument{
		Token:       request.Token,
		NewPassword: request.NewPassword,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(res)
}
