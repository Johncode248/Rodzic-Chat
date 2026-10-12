package childevents

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	childeventspg "github.com/Johncode248/Rodzic-Chat/project/common/childevents"
)

var ErrEventNotFound = errors.New("event not found")

func byId(id string) databasepg.Query {
	return databasepg.Query{Limit: 1, Groups: []databasepg.Group{
		{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: id}}},
	}}
}

// getEditable loads an event (with attendees) and checks accountId may change it.
func getEditable(ctx context.Context, accountId string, id string) (event childeventspg.ChildEvent, err error) {
	existing, err := gormdb.Select[childeventspg.ChildEvent](ctx, byId(id))
	if err != nil {
		return event, err
	}
	if len(existing) == 0 {
		return event, ErrEventNotFound
	}
	if err = LoadAttendees(ctx, existing); err != nil {
		return event, err
	}
	event = existing[0]

	allowed, err := canEdit(ctx, accountId, event)
	if err != nil {
		return event, err
	}
	if !allowed {
		return event, ErrNotYourChild
	}
	return event, nil
}

// Update applies a partial edit to the WHOLE event (for a repeating event: the whole series).
func Update(ctx context.Context, accountId string, id string, req childeventspg.UpdateRequest) (res childeventspg.ChildEvent, err error) {
	ctx = logger.Enter(ctx, "childevents.Update", accountId, id, req)
	defer logger.Exit(ctx, "childevents.Update", &res, &err)

	event, err := getEditable(ctx, accountId, id)
	if err != nil {
		return res, err
	}

	if req.Title != nil {
		if *req.Title == "" {
			return res, ErrTitleRequired
		}
		event.Title = *req.Title
	}
	if req.Description != nil {
		event.Description = *req.Description
	}
	if req.Location != nil {
		event.Location = *req.Location
	}
	if req.AllDay != nil {
		event.AllDay = *req.AllDay
	}
	if req.TimeZone != nil {
		event.TimeZone = validZone(*req.TimeZone)
	}
	if req.StartTime != nil {
		parsed, parseErr := parseTime(*req.StartTime)
		if parseErr != nil {
			return res, parseErr
		}
		event.StartTime = parsed
	}
	if req.EndTime != nil {
		if *req.EndTime == "" {
			event.EndTime = nil
		} else {
			parsed, parseErr := parseTime(*req.EndTime)
			if parseErr != nil {
				return res, parseErr
			}
			event.EndTime = &parsed
		}
	}
	if event.EndTime != nil && event.EndTime.Before(event.StartTime) {
		return res, ErrEndBeforeStart
	}

	freq, interval := event.Recurrence, event.RecurrenceInterval
	if req.Recurrence != nil {
		freq = *req.Recurrence
	}
	if req.RecurrenceInterval != nil {
		interval = *req.RecurrenceInterval
	}
	if freq, interval, err = validateRecurrence(freq, interval); err != nil {
		return res, err
	}
	event.Recurrence, event.RecurrenceInterval = freq, interval
	if freq == childeventspg.RecurrenceNone {
		event.RecurrenceUntil = nil
	} else if req.RecurrenceUntil != nil {
		if *req.RecurrenceUntil == "" {
			event.RecurrenceUntil = nil
		} else {
			parsed, parseErr := parseTime(*req.RecurrenceUntil)
			if parseErr != nil {
				return res, parseErr
			}
			event.RecurrenceUntil = &parsed
		}
	}
	if event.RecurrenceUntil != nil && event.RecurrenceUntil.Before(event.StartTime) {
		return res, ErrRecurrenceUntilBad
	}

	// Attendees: only touched when the request says so.
	if req.ChildIds != nil || req.ParentIds != nil || req.IncludeMe != nil {
		children := event.ChildIds
		if req.ChildIds != nil {
			children = *req.ChildIds
		}
		includeMe := false
		for _, p := range event.ParentIds {
			if p == accountId {
				includeMe = true
			}
		}
		if req.IncludeMe != nil {
			includeMe = *req.IncludeMe
		}
		// Unchanged co-parents stay valid; a new selection is validated against the caller.
		var others []string
		if req.ParentIds != nil {
			others = *req.ParentIds
		} else {
			for _, p := range event.ParentIds {
				if p != accountId {
					others = append(others, p)
				}
			}
		}
		newChildren, newParents, attErr := attendeesFrom(ctx, accountId, children, includeMe, others)
		if attErr != nil {
			return res, attErr
		}
		if err = replaceAttendees(ctx, event.Id, newChildren, newParents); err != nil {
			return res, err
		}
		event.ChildIds, event.ParentIds = newChildren, newParents
		event.ChildId = ""
		if len(newChildren) > 0 {
			event.ChildId = newChildren[0]
		}
	}
	event.UpdateTime = time.Now()

	if err = gormdb.Update(ctx, &event, byId(id)); err != nil {
		return res, err
	}
	return event, nil
}

// Delete removes an event (the whole series) -- same authorization as Update. The join rows go
// with it (ON DELETE CASCADE).
func Delete(ctx context.Context, accountId string, id string) (err error) {
	ctx = logger.Enter(ctx, "childevents.Delete", accountId, id)
	defer logger.Exit(ctx, "childevents.Delete", &err)

	if _, err = getEditable(ctx, accountId, id); err != nil {
		return err
	}

	_, err = gormdb.Delete[childeventspg.ChildEvent](ctx, byId(id))
	return err
}
