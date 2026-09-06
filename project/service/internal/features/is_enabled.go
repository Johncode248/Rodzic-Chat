package features

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	featurespg "github.com/Johncode248/Rodzic-Chat/project/common/features"
)

// IsEnabled reports whether the given feature is enabled for organizationId.
// Every organization should have an explicit OrganizationFeature row seeded
// on creation (see authorization.FinalizeRegisterMutation), so the "no row
// found" branch below is defensive only and defaults to disabled.
func IsEnabled(ctx context.Context, organizationId string, feature featurespg.Feature) (bool, error) {
	ctx = logger.Enter(ctx, "features.IsEnabled", organizationId, feature)
	var err error
	defer logger.Exit(ctx, "features.IsEnabled", &err)

	result, err := Select(ctx, featurespg.SelectArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "organization_id", Operator: "=", Value: organizationId},
						{Key: "feature", Operator: "=", Value: int(feature)},
					},
				},
			},
		},
	})
	if err != nil {
		return false, err
	}

	if len(result.OrganizationFeatures) == 0 {
		return false, nil
	}

	return result.OrganizationFeatures[0].Enabled, nil
}
