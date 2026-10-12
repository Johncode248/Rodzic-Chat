package childevents

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	childeventspg "github.com/Johncode248/Rodzic-Chat/project/common/childevents"
)

// ListForAccount returns every personal event relevant to accountId: one that involves any of
// the caller's children (childIds), one the caller attends themselves, or one they created --
// regardless of which linked parent created it. Attendees are loaded on each event.
func ListForAccount(ctx context.Context, accountId string, childIds []string) (res []childeventspg.ChildEvent, err error) {
	ctx = logger.Enter(ctx, "childevents.ListForAccount", accountId, childIds)
	defer logger.Exit(ctx, "childevents.ListForAccount", &res, &err)

	ids := map[string]bool{}
	if len(childIds) > 0 {
		links, selErr := gormdb.Select[childeventspg.ChildEventChild](ctx, inQuery("child_id", childIds))
		if selErr != nil {
			return nil, selErr
		}
		for _, l := range links {
			ids[l.EventId] = true
		}
	}
	attending, err := gormdb.Select[childeventspg.ChildEventParent](ctx, eq("account_id", accountId))
	if err != nil {
		return nil, err
	}
	for _, l := range attending {
		ids[l.EventId] = true
	}
	created, err := gormdb.Select[childeventspg.ChildEvent](ctx, eq("created_by_account_id", accountId))
	if err != nil {
		return nil, err
	}
	for _, e := range created {
		ids[e.Id] = true
	}
	if len(ids) == 0 {
		return nil, nil
	}

	all := make([]string, 0, len(ids))
	for id := range ids {
		all = append(all, id)
	}
	query := inQuery("id", all)
	query.OrderBy = databasepg.Order{Key: "start_time", Direction: databasepg.DirectionAscending}
	res, err = gormdb.Select[childeventspg.ChildEvent](ctx, query)
	if err != nil {
		return nil, err
	}
	if err = LoadAttendees(ctx, res); err != nil {
		return nil, err
	}
	return res, nil
}
