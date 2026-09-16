package events

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	eventspg "github.com/Johncode248/Rodzic-Chat/project/common/events"
)

func UpdateEvent(ctx context.Context, argument eventspg.EventUpdateArgument) (res eventspg.EventUpdateResponse, err error) {
	ctx = logger.Enter(ctx, "events.UpdateEvent", argument)
	defer logger.Exit(ctx, "events.UpdateEvent", &res, &err)

	var (
		usesMutation = &eventspg.EventUpdateMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = InitiateEventUpdateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	defer func() {
		if err != nil {
			if rollbackErr := RollbackEventUpdateMutation(ctx, usesMutation); rollbackErr != nil {
				panic(rollbackErr)
			}
		}
	}()

	if err = FinalizeEventUpdateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return eventspg.EventUpdateResponse{
		Event: usesMutation.Transient.Current,
	}, nil
}

func InitiateEventUpdateMutation(ctx context.Context, mutation *eventspg.EventUpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "events.InitiateEventUpdateMutation", mutation)
	defer logger.Exit(ctx, "events.InitiateEventUpdateMutation", &err)

	result, err := SelectEvents(ctx, eventspg.EventSelectArgument{Query: mutation.Foundation.Query})
	if err != nil {
		return err
	}

	if len(result.Events) == 0 {
		return nil
	}

	updateEvent := mutation.Foundation.Populate(result.Events[0])

	mutation.Transient = eventspg.EventTransient{
		Current: result.Events[0],
		Future:  updateEvent,
	}

	return nil
}

func FinalizeEventUpdateMutation(ctx context.Context, mutation *eventspg.EventUpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "events.FinalizeEventUpdateMutation", mutation)
	defer logger.Exit(ctx, "events.FinalizeEventUpdateMutation", &err)

	err = gormdb.Update(ctx, mutation.Transient.Future, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackEventUpdateMutation(ctx context.Context, mutation *eventspg.EventUpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "events.RollbackEventUpdateMutation", mutation)
	defer logger.Exit(ctx, "events.RollbackEventUpdateMutation", &err)

	if mutation.Transient.Current.Id == "" {
		return nil
	}

	return FinalizeEventUpdateMutation(ctx, &eventspg.EventUpdateMutation{
		Foundation: mutation.Foundation,
		Transient: eventspg.EventTransient{
			Current: mutation.Transient.Current,
			Future:  mutation.Transient.Previous,
		},
	})
}
