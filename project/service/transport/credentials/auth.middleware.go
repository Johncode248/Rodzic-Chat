package credentials

import (
	"context"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/credentials"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
			return
		}

		token, err := credentials.ExtractToken(context.Background(), authHeader)
		if err != nil {
			http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
			return
		}

		// A refresh token is signed with the same key and would otherwise
		// pass ExtractToken just fine -- without this check, a leaked
		// refresh token (30-day TTL, meant only for POST /public/refresh-token)
		// could be used directly as a full access token on every /private
		// route for up to 30 days instead of the intended 24h.
		if token.TokenKind != "access" {
			http.Error(w, "Invalid token type", http.StatusUnauthorized)
			return
		}

		// Access tokens are stateless, so a token minted for an account that
		// has since been deleted (or lives in another database) would still
		// verify and then fail deep inside a handler on a foreign key. Reject
		// it here with 401 so the client re-authenticates instead.
		accounts, err := gormdb.Select[accountspg.Account](r.Context(), databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: token.UserId}}},
			},
		})
		if err != nil {
			http.Error(w, "Error verifying account", http.StatusInternalServerError)
			return
		}
		if len(accounts) == 0 {
			http.Error(w, "Account no longer exists", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userId", token.UserId)
		ctx = context.WithValue(ctx, "userEmail", token.Email)
		ctx = context.WithValue(ctx, "organizationId", token.OrganizationId)
		ctx = context.WithValue(ctx, "accountType", token.AccountType)
		ctx = context.WithValue(ctx, "sessionId", token.SessionId)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
