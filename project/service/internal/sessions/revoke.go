package sessions

import (
	"context"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	sessionspg "github.com/Johncode248/Rodzic-Chat/project/common/sessions"
)

// RevokeById revokes a single session (e.g. "log out this device").
func RevokeById(ctx context.Context, sessionId string) (err error) {
	ctx = logger.Enter(ctx, "sessions.RevokeById", sessionId)
	defer logger.Exit(ctx, "sessions.RevokeById", &err)

	now := time.Now()

	_, err = Update(ctx, sessionspg.UpdateArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "id", Operator: "=", Value: sessionId},
					},
				},
			},
		},
		RevokedAt: &now,
	})

	return err
}

// RevokeAllForAccount revokes every not-yet-revoked session belonging to an
// account -- used on password reset and on account self-delete, since both
// are security events that should force every device to re-authenticate.
func RevokeAllForAccount(ctx context.Context, accountId string) (err error) {
	ctx = logger.Enter(ctx, "sessions.RevokeAllForAccount", accountId)
	defer logger.Exit(ctx, "sessions.RevokeAllForAccount", &err)

	result, err := Select(ctx, sessionspg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "account_id", Operator: "=", Value: accountId},
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	functions := []func() error{}
	for _, session := range result.Sessions {
		if session.RevokedAt != nil {
			continue
		}

		sessionId := session.Id
		functions = append(functions, func() error {
			return RevokeById(ctx, sessionId)
		})
	}

	return routines.Select(functions...)
}
