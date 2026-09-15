package authgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/authorization"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/oauthverify"
	"github.com/Johncode248/Rodzic-Chat/variables"
)

// LoginGoogleMobile handles POST /public/login/google -- Android only, the
// counterpart to Login (email+password) and LoginAppleMobile (iOS). The
// Android app performs the native Google Sign-In flow itself and hands us
// only the resulting ID token; this never sees the user's Google password.
func LoginGoogleMobile(w http.ResponseWriter, r *http.Request) {
	var (
		request authorizationpg.GoogleLoginRequest
		res     authorizationpg.LoginResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "authgt.LoginGoogleMobile")
	defer logger.Exit(ctx, "authgt.LoginGoogleMobile", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}
	if request.IdToken == "" {
		http.Error(w, "id_token is required", http.StatusBadRequest)
		return
	}

	email, err := oauthverify.VerifyGoogleIDToken(request.IdToken, variables.GoogleOAuthClientId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	result, err := authorization.LoginWithVerifiedEmail(ctx, authorizationpg.LoginProviderArgument{
		Email:      email,
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
