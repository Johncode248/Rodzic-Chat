package accountorganizations

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	accountorganizationspg "github.com/Johncode248/Rodzic-Chat/project/common/accountorganizations"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentchild"
)

// RemoveMember removes a parent's access to one organization: deletes the
// AccountOrganization row for (accountId, organizationId) and the
// ParentChild links tying that parent to children *in that organization
// specifically*, leaving every other organization's membership/links for
// that parent completely untouched. Callers should check afterwards (via
// Select) whether the account has any memberships left and cascade to a
// full account delete if not -- see
// project/service/transport/organizations/remove.member.go, which does
// that (kept out of this package to avoid an accountorganizations <->
// accounts import cycle: accounts already imports credentials, which
// resolves organization membership through this package).
func RemoveMember(ctx context.Context, organizationId, accountId string) (err error) {

	if _, err = Delete(ctx, accountorganizationspg.DeleteArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{
					{Key: "account_id", Operator: "=", Value: accountId},
					{Key: "organization_id", Operator: "=", Value: organizationId},
				}},
			},
		},
	}); err != nil {
		return err
	}

	childrenResult, err := children.Select(ctx, childrenpg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "organization_id", Operator: "=", Value: organizationId}}},
			},
		},
	})
	if err != nil {
		return err
	}

	if len(childrenResult.Children) == 0 {
		return nil
	}

	childIds := make([]string, 0, len(childrenResult.Children))
	for _, child := range childrenResult.Children {
		childIds = append(childIds, child.Id)
	}

	_, err = parentchild.Delete(ctx, parentchildpg.DeleteArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{
					{Key: "parent_id", Operator: "=", Value: accountId},
					{Key: "child_id", Operator: "IN", Value: childIds},
				}},
			},
		},
	})

	return err
}
