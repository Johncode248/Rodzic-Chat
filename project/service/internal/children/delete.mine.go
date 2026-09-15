package children

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	childeventspg "github.com/Johncode248/Rodzic-Chat/project/common/childevents"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
)

func byChildId(childId string) databasepg.Query {
	return databasepg.Query{Groups: []databasepg.Group{
		{Filters: []databasepg.Filter{{Key: "child_id", Operator: "=", Value: childId}}},
	}}
}

func byParentAndChild(accountId, childId string) databasepg.Query {
	return databasepg.Query{Groups: []databasepg.Group{
		{Filters: []databasepg.Filter{
			{Key: "parent_id", Operator: "=", Value: accountId},
			{Key: "child_id", Operator: "=", Value: childId},
		}},
	}}
}

// DeleteMine lets a parent remove childId from their own "Moje dzieci" list. What actually
// happens depends on whether the child is personal or placówka-connected (see Child's doc
// comment on OrganizationId):
//
//   - Personal child (OrganizationId == ""): fully deleted -- the Child row itself, every
//     ParentChild link to it (there may be a co-parent via parentinvites), and every personal
//     ChildEvent for it, since nothing in the app can meaningfully reference an orphaned personal
//     child once no parent manages it.
//   - Placówka-connected child: NOT deleted -- that would remove it from the school's own roster
//     and the other linked parent's view too, neither of which a single parent should be able to
//     unilaterally do. Only the caller's own ParentChild link is removed (see UnlinkMine); the
//     placówka's staff manage full removal from their own side.
func DeleteMine(ctx context.Context, accountId string, childId string) (err error) {
	ctx = logger.Enter(ctx, "children.DeleteMine", accountId, childId)
	defer logger.Exit(ctx, "children.DeleteMine", &err)

	child, err := getMineOwned(ctx, accountId, childId)
	if err != nil {
		return err
	}

	if child.OrganizationId != "" {
		return UnlinkMine(ctx, accountId, childId)
	}

	if _, err = gormdb.Delete[parentchildpg.ParentChild](ctx, byChildId(childId)); err != nil {
		return err
	}
	if _, err = gormdb.Delete[childeventspg.ChildEvent](ctx, byChildId(childId)); err != nil {
		return err
	}
	if _, err = gormdb.Delete[childrenpg.Child](ctx, byIdQuery(childId)); err != nil {
		return err
	}

	return nil
}

// UnlinkMine removes only the caller's own parent-child link to childId, leaving the Child row
// (and any co-parent's link to it) untouched -- see DeleteMine's doc comment for when this runs
// instead of a full delete.
func UnlinkMine(ctx context.Context, accountId string, childId string) (err error) {
	ctx = logger.Enter(ctx, "children.UnlinkMine", accountId, childId)
	defer logger.Exit(ctx, "children.UnlinkMine", &err)

	_, err = gormdb.Delete[parentchildpg.ParentChild](ctx, byParentAndChild(accountId, childId))
	return err
}
