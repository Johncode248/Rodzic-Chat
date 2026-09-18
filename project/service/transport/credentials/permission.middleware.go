package credentials

import (
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
)

// RequireAccountTypeOrPermission is like RequireAccountType, but also lets
// the request through when the caller's account (looked up fresh from the
// DB, since Permission grants aren't embedded in the access token and must
// take effect immediately, not after the caller's next login/refresh) has
// the given permission bit set -- e.g. a Teacher a director granted
// PermissionManagePayments to, without promoting them to CoDirector.
func RequireAccountTypeOrPermission(permission accountspg.Permission, types ...int) func(http.Handler) http.Handler {
	allowed := make(map[int]bool, len(types))
	for _, t := range types {
		allowed[t] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accountType, err := GetAccountTypeFromContext(r.Context())
			if err != nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if allowed[accountType] {
				next.ServeHTTP(w, r)
				return
			}

			userId, err := GetUserIdFromContext(r.Context())
			if err != nil || userId == "" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			result, err := accounts.Select(r.Context(), accountspg.SelectArgument{
				Query: databasepg.Query{
					Limit: 1,
					Groups: []databasepg.Group{
						{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: userId}}},
					},
				},
			})
			if err != nil || len(result.Accounts) == 0 || !result.Accounts[0].Permissions.Has(permission) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
