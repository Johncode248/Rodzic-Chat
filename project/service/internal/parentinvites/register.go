package parentinvites

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	credentialspg "github.com/Johncode248/Rodzic-Chat/project/common/credentials"
	parentinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/parentinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/credentials"
	"github.com/google/uuid"
)

var ErrInviteNotUsable = errors.New("this invite is expired or already used")
var ErrEmailAlreadyRegistered = errors.New("an account with this email already exists -- log in and use the invite link's \"already have an account\" option instead")

// Register creates a brand-new account for someone accepting a parent
// invite with no existing account (mirrors
// project/service/internal/authorization.RegisterParent -- same
// no-organization account shape), links it to every child the inviter
// parents, and logs it straight in.
func Register(ctx context.Context, token string, req parentinvitespg.RegisterRequest) (res authorizationpg.LoginReply, err error) {
	invite, err := getByToken(ctx, token)
	if err != nil {
		return res, err
	}
	if !invite.IsUsable(time.Now()) {
		return res, ErrInviteNotUsable
	}

	existing, err := accounts.Select(ctx, accountspg.SelectArgument{
		Query: databasepg.Query{Limit: 1, Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "login", Operator: "=", Value: req.Email}}},
		}},
	})
	if err != nil {
		return res, err
	}
	if len(existing.Accounts) > 0 {
		return res, ErrEmailAlreadyRegistered
	}

	accountId := uuid.NewString()
	if _, err = accounts.Create(ctx, accountspg.CreateArgument{
		Id:          accountId,
		Name:        req.Name,
		Login:       req.Email,
		Email:       req.Email,
		AccountType: accountspg.AccountTypeParent,
		Status:      accountspg.StatusEnabled,
	}); err != nil {
		return res, err
	}

	if _, err = credentials.Create(ctx, credentialspg.CreateArgument{
		UserId:   accountId,
		Password: req.Password,
	}); err != nil {
		return res, err
	}

	if _, err = linkToInviterChildren(ctx, invite, accountId); err != nil {
		return res, err
	}

	issued, err := credentials.IssueTokens(ctx, credentials.IssueTokensArgument{
		AccountId: accountId,
		Email:     req.Email,
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
