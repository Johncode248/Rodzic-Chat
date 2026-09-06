package events

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	eventspg "github.com/Johncode248/Rodzic-Chat/project/common/events"
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

	existing, err := SelectRegistrations(ctx, eventspg.RegistrationSelectArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "event_id", Operator: "=", Value: argument.EventId},
						{Key: "child_id", Operator: "=", Value: argument.ChildId},
					},
				},
			},
		},
	})
	if err != nil {
		return res, err
	}

	if len(existing.EventRegistrations) == 0 {
		return CreateRegistration(ctx, argument)
	}

	status := argument.Status
	consent := argument.ConsentGiven
	updateResult, err := UpdateRegistration(ctx, eventspg.RegistrationUpdateArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "id", Operator: "=", Value: existing.EventRegistrations[0].Id},
					},
				},
			},
		},
		ConfirmedBy:  argument.ConfirmedBy,
		Status:       &status,
		ConsentGiven: &consent,
	})
	if err != nil {
		return res, err
	}

	return eventspg.RegistrationCreateResponse{
		EventRegistration: updateResult.EventRegistration,
	}, nil
}
