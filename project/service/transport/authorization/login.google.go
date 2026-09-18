package authgt

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/Johncode248/Rodzic-Chat/logger"
	"github.com/Johncode248/Rodzic-Chat/variables"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOAuthConfig = &oauth2.Config{
	ClientID:     variables.GoogleOAuthClientId,
	ClientSecret: variables.GoogleOAuthClientSecret,
	RedirectURL:  fmt.Sprintf("%s%s", variables.BackendURL, "/login/callback"),
	Scopes: []string{
		"openid",
		"email",
		"profile",
	},
	Endpoint: google.Endpoint,
}

// googleOAuthStateCookie carries a random per-flow nonce so
// GoogleCallbackLoginHandler can verify the callback it's handling was
// actually initiated by the same browser (standard OAuth login-CSRF
// mitigation) -- without this, an attacker can start the flow themselves,
// capture the resulting redirect (code tied to the ATTACKER's own Google
// account), and trick a victim's browser into completing it, silently
// logging the victim into the attacker's account.
const googleOAuthStateCookie = "google_oauth_state"

func generateOAuthState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func LoginGoogle(w http.ResponseWriter, r *http.Request) {
	var (
		err error
	)

	ctx := logger.Enter(r.Context(), "authgt.LoginGoogle")
	defer logger.Exit(ctx, "authgt.LoginGoogle", &err)

	state, err := generateOAuthState()
	if err != nil {
		http.Error(w, "Error generating state", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     googleOAuthStateCookie,
		Value:    state,
		Path:     "/",
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	url := googleOAuthConfig.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
	)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)

	return
}
