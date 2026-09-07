package organizations

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
)

// SetSubscriptionFreeOverride is the only way SubscriptionFreeOverride ever
// changes -- gated at the transport layer by the platform-admin shared
// secret (see transport/organizations/subscription.override.go), never
// reachable through the director-facing Update endpoint.
func SetSubscriptionFreeOverride(ctx context.Context, organizationId string, freeOverride bool) (res organizationspg.Organization, err error) {
	ctx = logger.Enter(ctx, "organizations.SetSubscriptionFreeOverride", organizationId, freeOverride)
	defer logger.Exit(ctx, "organizations.SetSubscriptionFreeOverride", &res, &err)

	query := databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: organizationId}}},
		},
	}

	result, err := Select(ctx, organizationspg.SelectArgument{Query: query})
	if err != nil {
		return res, err
	}
	if len(result.Organizations) == 0 {
		return res, errors.New("organization not found")
	}

	organization := result.Organizations[0]
	organization.SubscriptionFreeOverride = freeOverride
	organization.UpdateTime = time.Now()

	if err = gormdb.Update(ctx, &organization, databasepg.Query{
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: organizationId}}},
		},
	}); err != nil {
		return res, err
	}

	return organization, nil
}
