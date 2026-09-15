package authgt

import (
	"fmt"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	"github.com/Johncode248/Rodzic-Chat/variables"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOAuthConfig2 = &oauth2.Config{
	ClientID:     variables.GoogleOAuthClientId,
	ClientSecret: variables.GoogleOAuthClientSecret,
	RedirectURL:  fmt.Sprintf("%s%s", variables.DOMEIN_B, "/register/callback"),
	Scopes: []string{
		"openid",
		"email",
		"profile",
	},
	Endpoint: google.Endpoint,
}

func RegisterGoogle(w http.ResponseWriter, r *http.Request) {
	var (
		err error
	)

	ctx := logger.Enter(r.Context(), "authgt.RegisterGoogle")
	defer logger.Exit(ctx, "authgt.RegisterGoogle", &err)

	http.Error(w, "app in dev mode", http.StatusInternalServerError)
	return

	url := googleOAuthConfig2.AuthCodeURL(
		"register",
		oauth2.AccessTypeOffline,
	)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)

	return
}
