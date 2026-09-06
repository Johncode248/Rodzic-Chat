package events

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	eventspg "github.com/Johncode248/Rodzic-Chat/project/common/events"
)

func SelectRegistrations(ctx context.Context, argument eventspg.RegistrationSelectArgument) (res eventspg.RegistrationSelectResponse, err error) {
	ctx = logger.Enter(ctx, "events.SelectRegistrations", argument)
	defer logger.Exit(ctx, "events.SelectRegistrations", &res, &err)

	var (
		usesMutation = &eventspg.RegistrationSelectMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = ProcessRegistrationSelectMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return eventspg.RegistrationSelectResponse{
		EventRegistrations: usesMutation.Transient,
	}, nil
}

func ProcessRegistrationSelectMutation(ctx context.Context, mutation *eventspg.RegistrationSelectMutation) (err error) {
	ctx = logger.Enter(ctx, "events.ProcessRegistrationSelectMutation", mutation)
	defer logger.Exit(ctx, "events.ProcessRegistrationSelectMutation", &err)

	result, err := gormdb.Select[eventspg.EventRegistration](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = result

	return nil
}
