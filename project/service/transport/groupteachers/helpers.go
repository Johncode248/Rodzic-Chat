package groupteachersgt

import (
	"context"
	"errors"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	groupspg "github.com/Johncode248/Rodzic-Chat/project/common/groups"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/groups"
)

var errGroupNotInOrganization = errors.New("group does not belong to the caller's organization")
var errTeacherNotInOrganization = errors.New("teacher account does not belong to the caller's organization")

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

// groupIdsInOrganization returns every group id in organizationId, used to
// scope a group-teacher list query (which has no organization_id column of
// its own) when the caller doesn't pass ?group_id=.
func groupIdsInOrganization(ctx context.Context, organizationId string) ([]string, error) {
	result, err := groups.Select(ctx, groupspg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "organization_id", Operator: "=", Value: organizationId}}},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(result.Groups))
	for _, group := range result.Groups {
		ids = append(ids, group.Id)
	}

	return ids, nil
}

func teacherInOrganization(ctx context.Context, teacherId, organizationId string) (bool, error) {
	if teacherId == "" || organizationId == "" {
		return false, nil
	}

	result, err := accounts.Select(ctx, accountspg.SelectArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "id", Operator: "=", Value: teacherId},
						{Key: "organization_id", Operator: "=", Value: organizationId},
						{Key: "account_type", Operator: "=", Value: accountspg.AccountTypeTeacher.Ordinal()},
					},
				},
			},
		},
	})
	if err != nil {
		return false, err
	}

	return len(result.Accounts) > 0, nil
}
