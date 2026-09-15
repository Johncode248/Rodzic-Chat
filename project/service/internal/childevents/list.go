package childevents

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	childeventspg "github.com/Johncode248/Rodzic-Chat/project/common/childevents"
)

// ListForChildren returns every personal event for the given child ids
// (the caller's own children -- see calendar.Get, which resolves that list
// via children.SelectMine before calling this), regardless of which linked
// parent created each one.
func ListForChildren(ctx context.Context, childIds []string) (res []childeventspg.ChildEvent, err error) {
	ctx = logger.Enter(ctx, "childevents.ListForChildren", childIds)
	defer logger.Exit(ctx, "childevents.ListForChildren", &res, &err)

	if len(childIds) == 0 {
		return nil, nil
	}

	return gormdb.Select[childeventspg.ChildEvent](ctx, databasepg.Query{
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "child_id", Operator: "IN", Value: childIds}}},
		},
		OrderBy: databasepg.Order{Key: "start_time", Direction: databasepg.DirectionAscending},
	})
}
