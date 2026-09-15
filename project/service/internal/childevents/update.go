package childevents

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	childeventspg "github.com/Johncode248/Rodzic-Chat/project/common/childevents"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentchild"
)

var ErrEventNotFound = errors.New("event not found")

func byId(id string) databasepg.Query {
	return databasepg.Query{Limit: 1, Groups: []databasepg.Group{
		{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: id}}},
	}}
}

func getOwned(ctx context.Context, accountId string, id string) (event childeventspg.ChildEvent, err error) {
	existing, err := gormdb.Select[childeventspg.ChildEvent](ctx, byId(id))
	if err != nil {
		return event, err
	}
	if len(existing) == 0 {
		return event, ErrEventNotFound
	}
	event = existing[0]

	isParent, err := parentchild.IsParent(ctx, accountId, event.ChildId)
	if err != nil {
		return event, err
	}
	if !isParent {
		return event, ErrNotYourChild
	}

	return event, nil
}

// Update applies a partial edit -- accountId must be a parent of the
// event's child (any linked parent may edit any event for that child, not
// just the one who created it).
func Update(ctx context.Context, accountId string, id string, req childeventspg.UpdateRequest) (res childeventspg.ChildEvent, err error) {
	ctx = logger.Enter(ctx, "childevents.Update", accountId, id, req)
	defer logger.Exit(ctx, "childevents.Update", &res, &err)

	event, err := getOwned(ctx, accountId, id)
	if err != nil {
		return res, err
	}

	if req.Title != nil {
		event.Title = *req.Title
	}
	if req.Description != nil {
		event.Description = *req.Description
	}
	if req.Location != nil {
		event.Location = *req.Location
	}
	if req.StartTime != nil {
		parsed, parseErr := time.Parse(time.RFC3339, *req.StartTime)
		if parseErr != nil {
			return res, parseErr
		}
		event.StartTime = parsed
	}
	if req.EndTime != nil {
		if *req.EndTime == "" {
			event.EndTime = nil
		} else {
			parsed, parseErr := time.Parse(time.RFC3339, *req.EndTime)
			if parseErr != nil {
				return res, parseErr
			}
			event.EndTime = &parsed
		}
	}
	event.UpdateTime = time.Now()

	if err = gormdb.Update(ctx, &event, byId(id)); err != nil {
		return res, err
	}

	return event, nil
}

// Delete removes a personal event -- same authorization as Update.
func Delete(ctx context.Context, accountId string, id string) (err error) {
	ctx = logger.Enter(ctx, "childevents.Delete", accountId, id)
	defer logger.Exit(ctx, "childevents.Delete", &err)

	if _, err = getOwned(ctx, accountId, id); err != nil {
		return err
	}

	_, err = gormdb.Delete[childeventspg.ChildEvent](ctx, byId(id))
	return err
}
