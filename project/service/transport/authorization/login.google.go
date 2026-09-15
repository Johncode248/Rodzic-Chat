package authgt

import (
	"fmt"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	"github.com/Johncode248/Rodzic-Chat/variables"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOAuthConfig = &oauth2.Config{
	ClientID:     variables.GoogleOAuthClientId,
	ClientSecret: variables.GoogleOAuthClientSecret,
	RedirectURL:  fmt.Sprintf("%s%s", variables.DOMEIN_B, "/login/callback"),
	Scopes: []string{
		"openid",
		"email",
		"profile",
	},
	Endpoint: google.Endpoint,
}

func LoginGoogle(w http.ResponseWriter, r *http.Request) {
	var (
		err error
	)

	ctx := logger.Enter(r.Context(), "authgt.LoginGoogle")
	defer logger.Exit(ctx, "authgt.LoginGoogle", &err)

	url := googleOAuthConfig.AuthCodeURL(
		"login",
		oauth2.AccessTypeOffline,
	)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)

	return
}
