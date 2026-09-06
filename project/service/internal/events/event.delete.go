package events

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	eventspg "github.com/Johncode248/Rodzic-Chat/project/common/events"
)

func DeleteEvent(ctx context.Context, argument eventspg.EventDeleteArgument) (res eventspg.EventDeleteResponse, err error) {
	ctx = logger.Enter(ctx, "events.DeleteEvent", argument)
	defer logger.Exit(ctx, "events.DeleteEvent", &res, &err)

	var (
		usesMutation = &eventspg.EventDeleteMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = InitiateEventDeleteMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	defer func() {
		if err != nil {
			if err = RollbackEventDeleteMutation(ctx, usesMutation); err != nil {
				panic(err)
			}
		}
	}()

	if err = FinalizeEventDeleteMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	for _, value := range usesMutation.Transient {
		res.Events = append(res.Events, value.Previous)
	}

	return res, nil
}

func InitiateEventDeleteMutation(ctx context.Context, mutation *eventspg.EventDeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "events.InitiateEventDeleteMutation", mutation)
	defer logger.Exit(ctx, "events.InitiateEventDeleteMutation", &err)

	result, err := SelectEvents(ctx, eventspg.EventSelectArgument{
		Query: mutation.Foundation.Query,
	})
	if err != nil {
		return err
	}

	if len(result.Events) == 0 {
		return nil
	}

	for _, event := range result.Events {
		mutation.Transient = append(mutation.Transient, eventspg.EventTransient{
			Current: event,
		})
	}

	return nil
}

func FinalizeEventDeleteMutation(ctx context.Context, mutation *eventspg.EventDeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "events.FinalizeEventDeleteMutation", mutation)
	defer logger.Exit(ctx, "events.FinalizeEventDeleteMutation", &err)

	_, err = gormdb.Delete[eventspg.Event](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	for index := range mutation.Transient {
		mutation.Transient[index] = mutation.Transient[index].Update()
	}

	return nil
}

func RollbackEventDeleteMutation(ctx context.Context, mutation *eventspg.EventDeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "events.RollbackEventDeleteMutation", mutation)
	defer logger.Exit(ctx, "events.RollbackEventDeleteMutation", &err)

	return routines.Select(
		func() error {
			functions := []func() error{}
			for _, transient := range mutation.Transient {
				if transient.Previous.Id == "" {
					continue
				}

				transient := transient
				functions = append(functions, func() error {
					return FinalizeEventCreateMutation(ctx, &eventspg.EventCreateMutation{
						Transient: eventspg.EventTransient{
							Future: transient.Previous,
						},
					})
				})
			}
			return routines.Select(functions...)
		},
	)
}
