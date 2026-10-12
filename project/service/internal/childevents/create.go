package childevents

import (
	"context"
	"errors"
	"time"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	childeventspg "github.com/Johncode248/Rodzic-Chat/project/common/childevents"
	"github.com/google/uuid"
)

var ErrNotYourChild = errors.New("you are not a parent of this child")
var ErrTitleRequired = errors.New("title is required")
var ErrStartTimeRequired = errors.New("start_time is required")

// Create adds a personal calendar event. The caller must be a parent of every listed child and
// may share it with a co-parent (see attendeesFrom); there is no organization/staff concept.
func Create(ctx context.Context, accountId string, req childeventspg.CreateRequest) (res childeventspg.ChildEvent, err error) {
	ctx = logger.Enter(ctx, "childevents.Create", accountId, req)
	defer logger.Exit(ctx, "childevents.Create", &res, &err)

	if req.Title == "" {
		return res, ErrTitleRequired
	}
	if req.StartTime == "" {
		return res, ErrStartTimeRequired
	}

	childIds := req.ChildIds
	if req.ChildId != "" {
		childIds = append([]string{req.ChildId}, childIds...)
	}
	includeMe := req.IncludeMe == nil || *req.IncludeMe
	children, parents, err := attendeesFrom(ctx, accountId, childIds, includeMe, req.ParentIds)
	if err != nil {
		return res, err
	}

	startTime, err := parseTime(req.StartTime)
	if err != nil {
		return res, err
	}
	var endTime *time.Time
	if req.EndTime != "" {
		parsed, parseErr := parseTime(req.EndTime)
		if parseErr != nil {
			return res, parseErr
		}
		if parsed.Before(startTime) {
			return res, ErrEndBeforeStart
		}
		endTime = &parsed
	}

	freq, interval, err := validateRecurrence(req.Recurrence, req.RecurrenceInterval)
	if err != nil {
		return res, err
	}
	var until *time.Time
	if freq != childeventspg.RecurrenceNone && req.RecurrenceUntil != "" {
		parsed, parseErr := parseTime(req.RecurrenceUntil)
		if parseErr != nil {
			return res, parseErr
		}
		if parsed.Before(startTime) {
			return res, ErrRecurrenceUntilBad
		}
		until = &parsed
	}

	primaryChild := ""
	if len(children) > 0 {
		primaryChild = children[0]
	}
	now := time.Now()
	event := childeventspg.ChildEvent{
		Id:                 uuid.NewString(),
		ChildId:            primaryChild,
		CreatedByAccountId: accountId,
		Title:              req.Title,
		Description:        req.Description,
		Location:           req.Location,
		StartTime:          startTime,
		EndTime:            endTime,
		AllDay:             req.AllDay,
		Recurrence:         freq,
		TimeZone:           validZone(req.TimeZone),
		RecurrenceInterval: interval,
		RecurrenceUntil:    until,
		CreateTime:         now,
		UpdateTime:         now,
	}
	if err = gormdb.Create(ctx, &event); err != nil {
		return res, err
	}
	if err = replaceAttendees(ctx, event.Id, children, parents); err != nil {
		return res, err
	}

	event.ChildIds = children
	event.ParentIds = parents
	return event, nil
}
