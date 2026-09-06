package authorization

import (
	"context"
	"errors"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	accountorganizationspg "github.com/Johncode248/Rodzic-Chat/project/common/accountorganizations"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accountorganizations"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/credentials"
)

var ErrNotAMember = errors.New("account is not a member of this organization")

// SwitchOrganization verifies the caller actually belongs to the requested
// organization (via AccountOrganization, falling back to Account.OrganizationId
// for accounts with no membership rows yet -- see resolveOrganizations'
// identical fallback), then re-mints tokens for the caller's existing
// session with that organization active.
func SwitchOrganization(ctx context.Context, argument authorizationpg.SwitchOrganizationArgument) (res authorizationpg.SwitchOrganizationResponse, err error) {

	if err = argument.Validate(); err != nil {
		return res, err
	}

	memberships, err := accountorganizations.Select(ctx, accountorganizationspg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "account_id", Operator: "=", Value: argument.AccountId}}},
			},
		},
	})
	if err != nil {
		return res, err
	}

	allowed := false
	if len(memberships.AccountOrganizations) == 0 {
		accountResult, err := accounts.Select(ctx, accountspg.SelectArgument{
			Query: databasepg.Query{
				Limit:  1,
				Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: argument.AccountId}}}},
			},
		})
		if err != nil {
			return res, err
		}
		allowed = len(accountResult.Accounts) > 0 && accountResult.Accounts[0].OrganizationId == argument.OrganizationId
	} else {
		for _, membership := range memberships.AccountOrganizations {
			if membership.OrganizationId == argument.OrganizationId {
				allowed = true
				break
			}
		}
	}

	if !allowed {
		return res, ErrNotAMember
	}

	issued, err := credentials.ReissueForSession(ctx, argument.SessionId, argument.AccountId, argument.Email, argument.OrganizationId)
	if err != nil {
		return res, err
	}

	return authorizationpg.SwitchOrganizationResponse{
		AccessToken:  issued.AccessToken,
		RefreshToken: issued.RefreshToken,
		ExpiresIn:    int(credentials.AccessTokenTTL.Seconds()),
	}, nil
}
