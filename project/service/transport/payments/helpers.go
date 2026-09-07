package paymentsgt

import (
	"context"
	"errors"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/children"
)

var errChildNotInOrganization = errors.New("child does not belong to the caller's organization")
var errParentNotInOrganization = errors.New("parent account does not belong to the caller's organization")

// childInOrganization / parentInOrganization mirror the same-named helpers
// in transport/parentchild and transport/groupteachers -- every payment
// recipient must be validated the same way a parent-child link is, since a
// Payment is effectively "assign a debt to this parent, for this child".
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

func parentInOrganization(ctx context.Context, parentId, organizationId string) (bool, error) {
	if parentId == "" || organizationId == "" {
		return false, nil
	}

	result, err := accounts.Select(ctx, accountspg.SelectArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "id", Operator: "=", Value: parentId},
						{Key: "organization_id", Operator: "=", Value: organizationId},
						{Key: "account_type", Operator: "=", Value: accountspg.AccountTypeParent.Ordinal()},
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
