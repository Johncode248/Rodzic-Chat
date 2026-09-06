package credentials

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	accountorganizationspg "github.com/Johncode248/Rodzic-Chat/project/common/accountorganizations"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accountorganizations"
)

// resolveOrganizations picks the JWT's active OrganizationId and full
// OrganizationIds list for an account. requestedOrganizationId (e.g. from
// "switch active organization") wins if the account is actually a member of
// it; otherwise the membership marked IsPrimary wins. If the account has no
// AccountOrganization rows at all yet (pre-backfill, or an environment that
// hasn't run it), this falls back to primaryOrganizationId (the account's
// plain Account.OrganizationId column) so behavior for every account stays
// unchanged until the membership table is populated. See pickActiveOrganization
// for the pure selection logic this wraps around the DB read.
func resolveOrganizations(ctx context.Context, accountId, primaryOrganizationId, requestedOrganizationId string) (active string, all []string, err error) {
	result, err := accountorganizations.Select(ctx, accountorganizationspg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "account_id", Operator: "=", Value: accountId}}},
			},
		},
	})
	if err != nil {
		return "", nil, err
	}

	memberships := make([]membership, 0, len(result.AccountOrganizations))
	for _, row := range result.AccountOrganizations {
		memberships = append(memberships, membership{OrganizationId: row.OrganizationId, IsPrimary: row.IsPrimary})
	}

	active, all = pickActiveOrganization(memberships, primaryOrganizationId, requestedOrganizationId)

	return active, all, nil
}

type membership struct {
	OrganizationId string
	IsPrimary      bool
}

// pickActiveOrganization is the pure decision logic behind resolveOrganizations,
// split out so it's unit-testable without a database:
//   - no memberships at all -> fall back to primaryOrganizationId (may be "")
//   - requestedOrganizationId is one of the memberships -> that one wins
//   - otherwise the membership marked IsPrimary wins
//   - otherwise (no membership marked primary -- shouldn't normally happen) the
//     first membership wins
func pickActiveOrganization(memberships []membership, primaryOrganizationId, requestedOrganizationId string) (active string, all []string) {
	if len(memberships) == 0 {
		if primaryOrganizationId == "" {
			return "", nil
		}
		return primaryOrganizationId, []string{primaryOrganizationId}
	}

	primary := ""
	member := make(map[string]bool, len(memberships))
	all = make([]string, 0, len(memberships))
	for _, m := range memberships {
		all = append(all, m.OrganizationId)
		member[m.OrganizationId] = true
		if m.IsPrimary {
			primary = m.OrganizationId
		}
	}

	if requestedOrganizationId != "" && member[requestedOrganizationId] {
		return requestedOrganizationId, all
	}

	if primary != "" {
		return primary, all
	}

	return all[0], all
}
