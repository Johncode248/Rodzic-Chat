package events

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	eventspg "github.com/Johncode248/Rodzic-Chat/project/common/events"
)

func SelectEvents(ctx context.Context, argument eventspg.EventSelectArgument) (res eventspg.EventSelectResponse, err error) {
	ctx = logger.Enter(ctx, "events.SelectEvents", argument)
	defer logger.Exit(ctx, "events.SelectEvents", &res, &err)

	var (
		usesMutation = &eventspg.EventSelectMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = ProcessEventSelectMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return eventspg.EventSelectResponse{
		Events: usesMutation.Transient,
	}, nil
}

func ProcessEventSelectMutation(ctx context.Context, mutation *eventspg.EventSelectMutation) (err error) {
	ctx = logger.Enter(ctx, "events.ProcessEventSelectMutation", mutation)
	defer logger.Exit(ctx, "events.ProcessEventSelectMutation", &err)

	result, err := gormdb.Select[eventspg.Event](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = result

	return nil
}
