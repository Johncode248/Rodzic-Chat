package authgt

import (
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
)

// GoogleCallbackRegisterHandler is intentionally disabled: /register now only
// creates a director + a brand-new organization (see register.go), and
// there's no way for a Google OAuth callback to supply organization_name/
// type/subdomain. Parent/teacher accounts are never self-registered at all —
// they're provisioned by a director via POST /private/accounts/invite. If
// "sign up with Google" for directors is wanted later, this handler should
// collect the organization fields (e.g. a form after the OAuth redirect)
// before calling authorization.Register.
func GoogleCallbackRegisterHandler(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Enter(r.Context(), "authgt.GoogleCallbackRegisterHandler")
	defer logger.Exit(ctx, "authgt.GoogleCallbackRegisterHandler")
	_ = ctx

	http.Error(w, "Google sign-up is not available yet — register a placówka via /public/register", http.StatusNotImplemented)
}
