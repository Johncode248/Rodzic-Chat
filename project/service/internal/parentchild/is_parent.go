package parentchild

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
)

// IsParent reports whether accountId is linked to childId as a parent --
// the authorization check every child-scoped, parent-facing endpoint needs
// (personal calendar events, editing a child's own fields, revealing its
// PESEL, ...). Small and shared rather than re-querying parentchild ad hoc
// at each call site.
func IsParent(ctx context.Context, accountId string, childId string) (bool, error) {
	res, err := Select(ctx, parentchildpg.SelectArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{
					{Key: "parent_id", Operator: "=", Value: accountId},
					{Key: "child_id", Operator: "=", Value: childId},
				}},
			},
		},
	})
	if err != nil {
		return false, err
	}
	return len(res.ParentChildren) > 0, nil
}
