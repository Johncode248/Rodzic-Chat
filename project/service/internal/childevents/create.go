package childevents

import (
	"context"
	"errors"
	"time"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	childeventspg "github.com/Johncode248/Rodzic-Chat/project/common/childevents"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentchild"
	"github.com/google/uuid"
)

var ErrNotYourChild = errors.New("you are not a parent of this child")
var ErrTitleRequired = errors.New("title is required")
var ErrStartTimeRequired = errors.New("start_time is required")

// Create adds a personal calendar event for a child -- accountId must be
// one of that child's linked parents (checked here, not left to the
// caller), since this has no other access control (no organization/staff
// concept applies to personal events).
func Create(ctx context.Context, accountId string, req childeventspg.CreateRequest) (res childeventspg.ChildEvent, err error) {
	ctx = logger.Enter(ctx, "childevents.Create", accountId, req)
	defer logger.Exit(ctx, "childevents.Create", &res, &err)

	if req.Title == "" {
		return res, ErrTitleRequired
	}
	if req.StartTime == "" {
		return res, ErrStartTimeRequired
	}

	isParent, err := parentchild.IsParent(ctx, accountId, req.ChildId)
	if err != nil {
		return res, err
	}
	if !isParent {
		return res, ErrNotYourChild
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return res, err
	}

	var endTime *time.Time
	if req.EndTime != "" {
		parsed, parseErr := time.Parse(time.RFC3339, req.EndTime)
		if parseErr != nil {
			return res, parseErr
		}
		endTime = &parsed
	}

	now := time.Now()
	event := childeventspg.ChildEvent{
		Id:                 uuid.NewString(),
		ChildId:            req.ChildId,
		CreatedByAccountId: accountId,
		Title:              req.Title,
		Description:        req.Description,
		Location:           req.Location,
		StartTime:          startTime,
		EndTime:            endTime,
		CreateTime:         now,
		UpdateTime:         now,
	}
	if err = gormdb.Create(ctx, &event); err != nil {
		return res, err
	}

	return event, nil
}
