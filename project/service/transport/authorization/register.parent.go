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

// RegisterParent handles POST /public/register/parent -- a parent signing
// up entirely on their own, no director invite involved (mobile-only
// client-side; see RegisterParentArgument's doc comment). Same
// token-verification-before-argument pattern as
// registrationlinksgt.PublicRegister for the provider-token branches.
func RegisterParent(w http.ResponseWriter, r *http.Request) {
	var (
		request authorizationpg.RegisterParentRequest
		res     authorizationpg.LoginResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "authgt.RegisterParent", &request)
	defer logger.Exit(ctx, "authgt.RegisterParent", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	var verifiedProviderEmail string
	switch {
	case request.GoogleIdToken != "":
		verifiedProviderEmail, err = oauthverify.VerifyGoogleIDToken(request.GoogleIdToken, variables.GoogleOAuthClientId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	case request.AppleIdentityToken != "":
		verifiedProviderEmail, err = oauthverify.VerifyAppleIdentityToken(request.AppleIdentityToken, variables.AppleAudience)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}

	result, err := authorization.RegisterParent(ctx, authorizationpg.RegisterParentArgument{
		Name:                  request.Name,
		Email:                 request.Email,
		Password:              request.Password,
		VerifiedProviderEmail: verifiedProviderEmail,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res = result.Inherit()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
