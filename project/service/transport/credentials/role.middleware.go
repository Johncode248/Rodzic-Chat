package credentials

import "net/http"

// RequireAccountType returns a middleware that only allows requests through when the
// caller's accountType (injected into the context by AuthMiddleware) is one of the
// given allowed types. Types are the int Ordinal() values of accountspg.AccountType.
func RequireAccountType(types ...int) func(http.Handler) http.Handler {
	allowed := make(map[int]bool, len(types))
	for _, t := range types {
		allowed[t] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accountType, err := GetAccountTypeFromContext(r.Context())
			if err != nil || !allowed[accountType] {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
