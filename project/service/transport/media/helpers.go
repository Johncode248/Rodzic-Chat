package mediagt

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentchild"
)

// parentInGroup reports whether accountId (a parent) has at least one
// child actually enrolled in groupId -- mirrors chats.requireParentInGroup.
// Without this, any parent could pass an arbitrary ?group_id= belonging to
// another class in their own organization and read that group's entire
// photo/document gallery.
func parentInGroup(ctx context.Context, accountId, groupId string) (bool, error) {
	if accountId == "" || groupId == "" {
		return false, nil
	}

	childrenResult, err := children.Select(ctx, childrenpg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "group_id", Operator: "=", Value: groupId}}},
			},
		},
	})
	if err != nil {
		return false, err
	}

	childIds := make([]string, 0, len(childrenResult.Children))
	for _, child := range childrenResult.Children {
		childIds = append(childIds, child.Id)
	}
	if len(childIds) == 0 {
		return false, nil
	}

	linkResult, err := parentchild.Select(ctx, parentchildpg.SelectArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "parent_id", Operator: "=", Value: accountId},
						{Key: "child_id", Operator: "IN", Value: childIds},
					},
				},
			},
		},
	})
	if err != nil {
		return false, err
	}

	return len(linkResult.ParentChildren) > 0, nil
}
