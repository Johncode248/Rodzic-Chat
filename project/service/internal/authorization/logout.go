package authorization

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/logger"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/sessions"
)

// Logout revokes the single session identified by sessionId (the caller's
// own, read from their access token) -- "log out this device", not every
// device. See sessions.RevokeAllForAccount for the "log out everywhere"
// case used by password reset and account deletion.
func Logout(ctx context.Context, sessionId string) (err error) {
	ctx = logger.Enter(ctx, "authorization.Logout", sessionId)
	defer logger.Exit(ctx, "authorization.Logout", &err)

	return sessions.RevokeById(ctx, sessionId)
}
