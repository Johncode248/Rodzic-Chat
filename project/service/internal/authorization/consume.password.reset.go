package authorization

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	credentialspg "github.com/Johncode248/Rodzic-Chat/project/common/credentials"
	passwordresetspg "github.com/Johncode248/Rodzic-Chat/project/common/passwordresets"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/credentials"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/passwordresets"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/sessions"
)

// ConsumePasswordReset verifies a raw reset token (against its stored
// hash), sets the new password, marks the token used so it can't be
// replayed, and revokes every session on the account -- a password reset is
// exactly the security event that should force every device to
// re-authenticate.
func ConsumePasswordReset(ctx context.Context, argument authorizationpg.ConsumePasswordResetArgument) (err error) {
	ctx = logger.Enter(ctx, "authorization.ConsumePasswordReset", argument)
	defer logger.Exit(ctx, "authorization.ConsumePasswordReset", &err)

	if err = argument.Validate(); err != nil {
		return err
	}

	tokenHash := credentials.HashToken(argument.Token)

	result, err := passwordresets.Select(ctx, passwordresetspg.SelectArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "token_hash", Operator: "=", Value: tokenHash},
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	if len(result.PasswordResets) == 0 {
		return errors.New("invalid or expired reset token")
	}

	reset := result.PasswordResets[0]

	if reset.UsedAt != nil {
		return errors.New("reset token already used")
	}

	if reset.ExpiresAt.Before(time.Now()) {
		return errors.New("reset token expired")
	}

	if _, err = credentials.Update(ctx, credentialspg.UpdateArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "user_id", Operator: "=", Value: reset.AccountId},
					},
				},
			},
		},
		Password: &argument.NewPassword,
	}); err != nil {
		return err
	}

	now := time.Now()
	if _, err = passwordresets.Update(ctx, passwordresetspg.UpdateArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "id", Operator: "=", Value: reset.Id},
					},
				},
			},
		},
		UsedAt: &now,
	}); err != nil {
		return err
	}

	return sessions.RevokeAllForAccount(ctx, reset.AccountId)
}
