package authgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/authorization"
)

func Login(w http.ResponseWriter, r *http.Request) {
	var (
		request authorizationpg.LoginRequest
		res     authorizationpg.LoginResponse
		err     error
	)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	ctx := logger.Enter(r.Context(), "authgt.Login", &request)
	defer logger.Exit(ctx, "authgt.Login", &res, &err)

	result, err := authorization.Login(ctx, authorizationpg.LoginArgument{
		Login:      request.Login,
		Password:   request.Password,
		DeviceId:   request.DeviceId,
		DeviceName: request.DeviceName,
		UserAgent:  r.UserAgent(),
		IpAddress:  r.RemoteAddr,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	res = result.Inherit()
	json.NewEncoder(w).Encode(res)
}
