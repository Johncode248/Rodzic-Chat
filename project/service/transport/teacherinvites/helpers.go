package teacherinvitesgt

import (
	"context"
	"errors"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	groupspg "github.com/Johncode248/Rodzic-Chat/project/common/groups"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/groups"
)

var errGroupNotInOrganization = errors.New("group does not belong to the caller's organization")

// groupInOrganization returns true when groupId is empty (a teacher invite
// may be created without a group assigned yet) or when it exists within
// organizationId.
func groupInOrganization(ctx context.Context, groupId, organizationId string) (bool, error) {
	if groupId == "" {
		return true, nil
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
