package authgt

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/credentials"
)

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	var (
		request authorizationpg.RefreshTokenRequest
		res     authorizationpg.RefreshTokenResponse
		err     error
	)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	ctx := logger.Enter(r.Context(), "authgt.RefreshToken", &request)
	defer logger.Exit(ctx, "authgt.RefreshToken", &res, &err)

	result, err := credentials.RefreshToken(ctx, authorizationpg.RefreshTokenArgument{
		RefreshToken: request.RefreshToken,
	})
	if err != nil {
		// Only a token the server can positively say is invalid/expired/
		// revoked should tell the client "you're logged out" (401) -- any
		// other failure (DB hiccup, etc.) is 500 so the client's
		// refresh-and-retry logic knows NOT to wipe the session over it.
		if errors.Is(err, credentials.ErrInvalidRefreshToken) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	res = result.Inherit()
	json.NewEncoder(w).Encode(res)
	return
}
