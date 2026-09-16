package groupsgt

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
)

// teacherInOrganization verifies that teacherId belongs to organizationId
// and is actually a Teacher account, mirroring groupteachersgt's helper of
// the same name -- without this, a director could set a group's primary
// TeacherId to any account id (a parent, another organization's staff, or
// a nonexistent id).
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
