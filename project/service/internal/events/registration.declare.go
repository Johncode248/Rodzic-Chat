package events

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	eventspg "github.com/Johncode248/Rodzic-Chat/project/common/events"
	"gorm.io/gorm"
)

var ErrRegistrationDeadlinePassed = errors.New("registration deadline for this event has passed")

// Declare records a parent's "będzie / nie będzie" decision for a child on
// an event. It is an upsert keyed on (event_id, child_id): a repeated
// declaration updates the existing row (status/consent) instead of creating
// a duplicate. Rejects the declaration once the event's RegistrationDeadline
// (if any) is in the past.
func Declare(ctx context.Context, argument eventspg.RegistrationCreateArgument) (res eventspg.RegistrationCreateResponse, err error) {
	ctx = logger.Enter(ctx, "events.Declare", argument)
	defer logger.Exit(ctx, "events.Declare", &res, &err)

	eventResult, err := SelectEvents(ctx, eventspg.EventSelectArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "id", Operator: "=", Value: argument.EventId},
					},
				},
			},
		},
	})
	if err != nil {
		return res, err
	}
	if len(eventResult.Events) == 0 {
		return res, errors.New("event not found")
	}

	event := eventResult.Events[0]
	if event.RegistrationDeadline != nil && event.RegistrationDeadline.Before(time.Now()) {
		return res, ErrRegistrationDeadlinePassed
	}

	// The select-then-create-or-update sequence below has to happen inside
	// one DB transaction serialized per (event, child), otherwise two
	// concurrent declarations for the same child/event (a double-tapped
	// "będzie" button, or a client retrying a timed-out request) can both
	// see "no existing registration" and both insert, producing two rows
	// for what is meant to be a single upsert -- a classic TOCTOU race.
	// pg_advisory_xact_lock keyed on event_id+child_id (via hashtext)
	// makes every concurrent call for this pair queue up one at a time,
	// mirroring the pattern already used for the per-account channel limit
	// in chatinvites.Create.
	var registration eventspg.EventRegistration
	err = gormdb.GormDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", argument.EventId+":"+argument.ChildId).Error; err != nil {
			return err
		}

		var existing eventspg.EventRegistration
		findErr := tx.Where("event_id = ? AND child_id = ?", argument.EventId, argument.ChildId).First(&existing).Error
		switch {
		case errors.Is(findErr, gorm.ErrRecordNotFound):
			registration = argument.Populate()
			return tx.Create(&registration).Error
		case findErr != nil:
			return findErr
		default:
			existing.ConfirmedBy = argument.ConfirmedBy
			existing.Status = argument.Status
			existing.UpdateTime = time.Now()
			registration = existing
			return tx.Save(&registration).Error
		}
	})
	if err != nil {
		return res, err
	}

	return eventspg.RegistrationCreateResponse{EventRegistration: registration}, nil
}
