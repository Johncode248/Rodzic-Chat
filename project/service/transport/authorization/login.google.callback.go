package authgt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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

	http.Redirect(w, r, fmt.Sprintf("%s/dashboard?access_token=%s&refresh_token=%s", variables.DOMEIN_F, issued.AccessToken, issued.RefreshToken), http.StatusFound)
}
