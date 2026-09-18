package chats

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentchild"
)

// Recipients returns every account id that can read channel -- the push
// notification fan-out list for a new message in it, mirroring
// CheckReadPermission's rules rather than checkPostingPermission's (a
// parent who can only post in their own group's channels can still read an
// org-wide Announcement channel, for instance). Staff (director/
// co-director/teacher) can read every channel in their organization; a
// parent can read an org-wide channel (GroupId == "") or one scoped to a
// group their child is actually in.
func Recipients(ctx context.Context, channel chatspg.Channel) ([]string, error) {
	staffResult, err := accounts.Select(ctx, accountspg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{
				{Key: "organization_id", Operator: "=", Value: channel.OrganizationId},
				{Key: "account_type", Operator: "IN", Value: []int{
					accountspg.AccountTypeDirector.Ordinal(),
					accountspg.AccountTypeCoDirector.Ordinal(),
					accountspg.AccountTypeTeacher.Ordinal(),
				}},
			}}},
		},
	})
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(staffResult.Accounts))
	for _, a := range staffResult.Accounts {
		ids = append(ids, a.Id)
	}

	if channel.GroupId == "" {
		parentsResult, err := accounts.Select(ctx, accountspg.SelectArgument{
			Query: databasepg.Query{
				Groups: []databasepg.Group{{Filters: []databasepg.Filter{
					{Key: "organization_id", Operator: "=", Value: channel.OrganizationId},
					{Key: "account_type", Operator: "=", Value: accountspg.AccountTypeParent.Ordinal()},
				}}},
			},
		})
		if err != nil {
			return nil, err
		}
		for _, a := range parentsResult.Accounts {
			ids = append(ids, a.Id)
		}
		return ids, nil
	}

	childrenResult, err := children.Select(ctx, childrenpg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "group_id", Operator: "=", Value: channel.GroupId}}}},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(childrenResult.Children) == 0 {
		return ids, nil
	}
	childIds := make([]string, 0, len(childrenResult.Children))
	for _, c := range childrenResult.Children {
		childIds = append(childIds, c.Id)
	}

	linksResult, err := parentchild.Select(ctx, parentchildpg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "child_id", Operator: "IN", Value: childIds}}}},
		},
	})
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		seen[id] = true
	}
	for _, link := range linksResult.ParentChildren {
		if seen[link.ParentId] {
			continue
		}
		seen[link.ParentId] = true
		ids = append(ids, link.ParentId)
	}

	return ids, nil
}
