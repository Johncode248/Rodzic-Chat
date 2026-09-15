package authorization

import (
	"context"
	"errors"
	"strings"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	credentialspg "github.com/Johncode248/Rodzic-Chat/project/common/credentials"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/credentials"
	"github.com/google/uuid"
)

var ErrParentAccountAlreadyExists = errors.New("an account with this email already exists")

// RegisterParent creates a brand-new Parent account with no organization at
// all (see RegisterParentArgument's doc comment) and immediately logs it
// in, same as the director-only Register does. Unlike
// registrationlinks.PublicRegister, there is no existing-account-joining
// branch here -- this only ever creates a fresh account, since without an
// invite link there is no group/child context to attach to anyway.
func RegisterParent(ctx context.Context, argument authorizationpg.RegisterParentArgument) (res authorizationpg.LoginReply, err error) {
	if err = argument.Validate(); err != nil {
		return res, err
	}

	effectiveEmail := argument.EffectiveEmail()

	existing, err := accounts.Select(ctx, accountspg.SelectArgument{
		Query: databasepg.Query{
			Limit:  1,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "login", Operator: "=", Value: effectiveEmail}}}},
		},
	})
	if err != nil {
		return res, err
	}
	if len(existing.Accounts) > 0 {
		return res, ErrParentAccountAlreadyExists
	}

	accountId := uuid.NewString()

	// accountspg.CreateArgument requires a non-blank Name (shared by every account-creation
	// path in the app, so it isn't loosened just for this one). The mobile onboarding flow
	// collects the real name right after registering via PATCH /private/accounts, so a
	// placeholder here is only ever transient -- fall back to the email's local part rather
	// than leave the account genuinely nameless in the meantime.
	name := argument.Name
	if name == "" {
		name = strings.SplitN(effectiveEmail, "@", 2)[0]
	}

	if _, err = accounts.Create(ctx, accountspg.CreateArgument{
		Id:          accountId,
		Name:        name,
		Login:       effectiveEmail,
		Email:       effectiveEmail,
		AccountType: accountspg.AccountTypeParent,
		Status:      accountspg.StatusEnabled,
	}); err != nil {
		return res, err
	}

	if argument.Password != "" {
		if _, err = credentials.Create(ctx, credentialspg.CreateArgument{
			UserId:   accountId,
			Password: argument.Password,
		}); err != nil {
			return res, err
		}
	}

	issued, err := credentials.IssueTokens(ctx, credentials.IssueTokensArgument{
		AccountId: accountId,
		Email:     effectiveEmail,
	})
	if err != nil {
		return res, err
	}

	return authorizationpg.LoginReply{
		AccessToken:  issued.AccessToken,
		RefreshToken: issued.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(credentials.AccessTokenTTL.Seconds()),
		ExpiresAt:    issued.ExpiresAt.Unix(),
	}, nil
}
