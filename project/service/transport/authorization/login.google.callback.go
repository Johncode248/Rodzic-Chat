package authgt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/credentials"
	"github.com/Johncode248/Rodzic-Chat/variables"
)

func GoogleCallbackLoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Enter(r.Context(), "authgt.GoogleCallbackLoginHandler")
	defer logger.Exit(ctx, "authgt.GoogleCallbackLoginHandler")

	// Verify this callback belongs to a flow THIS browser started (see
	// LoginGoogle's doc comment on googleOAuthStateCookie) -- clear the
	// cookie either way, it's single-use.
	stateCookie, cookieErr := r.Cookie(googleOAuthStateCookie)
	http.SetCookie(w, &http.Cookie{
		Name:     googleOAuthStateCookie,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	if cookieErr != nil || stateCookie.Value == "" || r.URL.Query().Get("state") != stateCookie.Value {
		http.Error(w, "Invalid or missing OAuth state", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Brak code", http.StatusBadRequest)
		return
	}

	token, err := googleOAuthConfig.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "Token exchange error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := googleOAuthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Userinfo error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var user map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&user)

	// Google's userinfo v2 endpoint reports verified_email explicitly for a
	// reason -- an unverified address must never be trusted to match an
	// existing Account by email (that's exactly the "find by email, no
	// password check" trust LoginWithVerifiedEmail's doc comment describes,
	// and it only holds if the email really is verified).
	if verified, ok := user["verified_email"].(bool); !ok || !verified {
		http.Error(w, "Google account email is not verified", http.StatusForbidden)
		return
	}

	accountList, err := gormdb.Select[accountspg.Account](context.Background(), databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{
				Filters: []databasepg.Filter{
					{
						Key:      "email",
						Operator: "=",
						Value:    user["email"],
					},
				},
			},
		},
	})
	if err != nil || len(accountList) == 0 {
		http.Error(w, "account not found", http.StatusNotFound)
		return
	}

	issued, err := credentials.IssueTokens(ctx, credentials.IssueTokensArgument{
		AccountId: accountList[0].Id,
		Email:     accountList[0].Login,
		UserAgent: r.UserAgent(),
		IpAddress: r.RemoteAddr,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Tokens go in the URL fragment, not the query string: a fragment is
	// never sent to the server (or to any proxy/access log along the way)
	// and isn't included in the Referer header of requests the dashboard
	// page then makes -- unlike a query string, which ends up in browser
	// history, server logs, and any analytics script reading
	// location.href. The frontend must read window.location.hash and then
	// call history.replaceState to scrub it immediately.
	http.Redirect(w, r, fmt.Sprintf("%s/dashboard#access_token=%s&refresh_token=%s", variables.DOMEIN_F, issued.AccessToken, issued.RefreshToken), http.StatusFound)
}
