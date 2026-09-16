package authgt

import (
	"encoding/json"
	"log"
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

	// Always respond 200 with the same body whether or not request.Email
	// matches an account -- SendPasswordResetEmail already treats "no such
	// account" as a silent no-op rather than an error precisely so this
	// endpoint can't be used to enumerate registered emails by comparing
	// response codes (previously 200 vs 409).
	if sendErr := authorization.SendPasswordResetEmail(ctx, request.Email); sendErr != nil {
		log.Printf("authgt.PasswordRecovery: SendPasswordResetEmail failed: %v", sendErr)
	}

	json.NewEncoder(w).Encode(res)
}
