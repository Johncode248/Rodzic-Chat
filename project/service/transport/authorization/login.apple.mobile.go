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

// LoginAppleMobile handles POST /public/login/apple -- iOS only, the
// counterpart to Login (email+password) and LoginGoogleMobile (Android).
// The iOS app performs the native Sign in with Apple flow itself and hands
// us only the resulting identity token.
func LoginAppleMobile(w http.ResponseWriter, r *http.Request) {
	var (
		request authorizationpg.AppleLoginRequest
		res     authorizationpg.LoginResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "authgt.LoginAppleMobile")
	defer logger.Exit(ctx, "authgt.LoginAppleMobile", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}
	if request.IdentityToken == "" {
		http.Error(w, "identity_token is required", http.StatusBadRequest)
		return
	}

	email, err := oauthverify.VerifyAppleIdentityToken(request.IdentityToken, variables.AppleAudience)
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
