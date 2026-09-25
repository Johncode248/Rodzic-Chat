package authorization

import (
	"context"
	"errors"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	credentialspg "github.com/Johncode248/Rodzic-Chat/project/common/credentials"
	sessionspg "github.com/Johncode248/Rodzic-Chat/project/common/sessions"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/credentials"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/sessions"
	"golang.org/x/crypto/bcrypt"
)

// ErrWrongCurrentPassword / ErrNoPasswordSet are surfaced to the caller as-is (unlike Login's
// generic error there is no enumeration risk here: the caller is already authenticated).
var (
	ErrWrongCurrentPassword = errors.New("current password is incorrect")
	ErrNoPasswordSet        = errors.New("this account has no password (signed in with Google/Apple) -- use \"forgot password\" to set one")
	ErrSamePassword         = errors.New("new password must differ from the current one")
)

// ChangePassword lets a signed-in account replace its own password after proving it knows the
// current one. Every OTHER session of the account is revoked (a changed password is exactly the
// event that should sign other devices out, e.g. a lost phone) while the caller's own session
// stays valid so the app isn't kicked to the login screen.
func ChangePassword(ctx context.Context, argument authorizationpg.ChangePasswordArgument) (err error) {
	ctx = logger.Enter(ctx, "authorization.ChangePassword", authorizationpg.ChangePasswordArgument{AccountId: argument.AccountId})
	defer logger.Exit(ctx, "authorization.ChangePassword", &err)

	if err = argument.Validate(); err != nil {
		return err
	}

	credentialList, err := gormdb.Select[credentialspg.Credential](ctx, databasepg.Query{
		Limit:  1,
		Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "user_id", Operator: "=", Value: argument.AccountId}}}},
	})
	if err != nil {
		return err
	}
	if len(credentialList) == 0 || credentialList[0].Password == "" {
		return ErrNoPasswordSet
	}

	if bcrypt.CompareHashAndPassword([]byte(credentialList[0].Password), []byte(argument.CurrentPassword)) != nil {
		return ErrWrongCurrentPassword
	}
	if argument.CurrentPassword == argument.NewPassword {
		return ErrSamePassword
	}

	newPassword := argument.NewPassword
	if _, err = credentials.Update(ctx, credentialspg.UpdateArgument{
		Query:    databasepg.Query{Limit: 1, Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "user_id", Operator: "=", Value: argument.AccountId}}}}},
		UserId:   argument.AccountId,
		Password: &newPassword,
	}); err != nil {
		return err
	}

	// Revoke every other live session. Best effort ordering matters less than completing the
	// password change, but a failure here is still returned so the client doesn't claim success.
	result, err := sessions.Select(ctx, sessionspg.SelectArgument{
		Query: databasepg.Query{Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "account_id", Operator: "=", Value: argument.AccountId}}}}},
	})
	if err != nil {
		return err
	}
	functions := []func() error{}
	for _, session := range result.Sessions {
		if session.RevokedAt != nil || session.Id == argument.SessionId {
			continue
		}
		sessionId := session.Id
		functions = append(functions, func() error { return sessions.RevokeById(ctx, sessionId) })
	}
	return routines.Select(functions...)
}
