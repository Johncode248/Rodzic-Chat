package parentchildgt

import (
	"context"
	"errors"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/children"
)

// childInOrganization verifies that childId belongs to organizationId.
// ParentChild rows have no organization_id column of their own, so every
// write/read path that touches a specific child must check the child's
// organization this way to preserve the multi-tenancy invariant.
func childInOrganization(ctx context.Context, childId, organizationId string) (bool, error) {
	if childId == "" || organizationId == "" {
		return false, nil
	}

	result, err := children.Select(ctx, childrenpg.SelectArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "id", Operator: "=", Value: childId},
						{Key: "organization_id", Operator: "=", Value: organizationId},
					},
				},
			},
		},
	})
	if err != nil {
		return false, err
	}

	return len(result.Children) > 0, nil
}

// childIdsInOrganization returns the ids of every child in organizationId,
// used to scope ParentChild list queries (which have no organization_id
// column of their own) via a "child_id IN (...)" filter.
func childIdsInOrganization(ctx context.Context, organizationId string) ([]string, error) {
	result, err := children.Select(ctx, childrenpg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "organization_id", Operator: "=", Value: organizationId},
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(result.Children))
	for _, child := range result.Children {
		ids = append(ids, child.Id)
	}

	return ids, nil
}

var errChildNotInOrganization = errors.New("child does not belong to the caller's organization")
