package authgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/authorization"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

func Logout(w http.ResponseWriter, r *http.Request) {
	var (
		res authorizationpg.LogoutResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "authgt.Logout")
	defer logger.Exit(ctx, "authgt.Logout", &res, &err)

	sessionId, err := credentials.GetSessionIdFromContext(ctx)
	if err != nil {
		http.Error(w, "Auth went wrong", http.StatusForbidden)
		return
	}

	if err = authorization.Logout(ctx, sessionId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
