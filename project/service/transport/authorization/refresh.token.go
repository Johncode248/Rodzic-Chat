package authgt

import (
	"encoding/json"
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	res = result.Inherit()
	json.NewEncoder(w).Encode(res)
	return
}
