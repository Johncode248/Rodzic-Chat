package parentinvites

import (
	"context"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
	parentinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/parentinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentchild"
)

const maxParentsPerChild = 2

// linkToInviterChildren links accepterAccountId, as a parent, to every
// child inviterAccountId already parents -- skipping any child already at
// the two-parent cap (see parentchildpg.CreateArgument's usage in
// registrationlinks for the same rule) or one accepterAccountId is
// somehow already linked to. Marks invite consumed on success.
func linkToInviterChildren(ctx context.Context, invite parentinvitespg.ParentInvite, accepterAccountId string) (res parentinvitespg.AcceptResponse, err error) {
	ctx = logger.Enter(ctx, "parentinvites.linkToInviterChildren", invite.Id, accepterAccountId)
	defer logger.Exit(ctx, "parentinvites.linkToInviterChildren", &res, &err)

	inviterLinks, err := parentchild.Select(ctx, parentchildpg.SelectArgument{
		Query: databasepg.Query{Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "parent_id", Operator: "=", Value: invite.InviterAccountId}}},
		}},
	})
	if err != nil {
		return res, err
	}

	for _, link := range inviterLinks.ParentChildren {
		existingForChild, selectErr := parentchild.Select(ctx, parentchildpg.SelectArgument{
			Query: databasepg.Query{Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "child_id", Operator: "=", Value: link.ChildId}}},
			}},
		})
		if selectErr != nil {
			return res, selectErr
		}

		alreadyLinked := false
		for _, existing := range existingForChild.ParentChildren {
			if existing.ParentId == accepterAccountId {
				alreadyLinked = true
				break
			}
		}
		if alreadyLinked {
			res.LinkedChildIds = append(res.LinkedChildIds, link.ChildId)
			continue
		}
		if len(existingForChild.ParentChildren) >= maxParentsPerChild {
			res.SkippedChildIds = append(res.SkippedChildIds, link.ChildId)
			continue
		}

		if _, createErr := parentchild.Create(ctx, parentchildpg.CreateArgument{
			ParentId:  accepterAccountId,
			ChildId:   link.ChildId,
			Relation:  parentchildpg.RelationTypeGuardian,
			CanPickup: true,
		}); createErr != nil {
			return res, createErr
		}
		res.LinkedChildIds = append(res.LinkedChildIds, link.ChildId)
	}

	now := time.Now()
	invite.ConsumedByAccountId = accepterAccountId
	invite.ConsumedAt = &now
	if err = gormdb.Update(ctx, &invite, byToken(invite.Token)); err != nil {
		return res, err
	}

	return res, nil
}
