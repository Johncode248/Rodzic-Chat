package childrengt

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	groupspg "github.com/Johncode248/Rodzic-Chat/project/common/groups"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/groups"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentchild"
)

// groupInOrganization verifies that groupId belongs to organizationId,
// mirroring groupteachersgt's helper of the same name -- without this, a
// director could set a child's group_id to a group belonging to a
// different organization entirely.
func groupInOrganization(ctx context.Context, groupId, organizationId string) (bool, error) {
	if groupId == "" || organizationId == "" {
		return false, nil
	}

	result, err := groups.Select(ctx, groupspg.SelectArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "id", Operator: "=", Value: groupId},
						{Key: "organization_id", Operator: "=", Value: organizationId},
					},
				},
			},
		},
	})
	if err != nil {
		return false, err
	}

	return len(result.Groups) > 0, nil
}

// childIdsForParent returns every child id linked to accountId, mirroring
// absencesgt's helper of the same name -- used to scope a parent's own
// /children listing to only their own children.
func childIdsForParent(ctx context.Context, accountId string) ([]string, error) {
	result, err := parentchild.Select(ctx, parentchildpg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "parent_id", Operator: "=", Value: accountId},
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(result.ParentChildren))
	for _, pc := range result.ParentChildren {
		ids = append(ids, pc.ChildId)
	}

	return ids, nil
}
