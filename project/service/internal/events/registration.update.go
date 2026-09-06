package events

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	eventspg "github.com/Johncode248/Rodzic-Chat/project/common/events"
)

func UpdateRegistration(ctx context.Context, argument eventspg.RegistrationUpdateArgument) (res eventspg.RegistrationUpdateResponse, err error) {
	ctx = logger.Enter(ctx, "events.UpdateRegistration", argument)
	defer logger.Exit(ctx, "events.UpdateRegistration", &res, &err)

	var (
		usesMutation = &eventspg.RegistrationUpdateMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = InitiateRegistrationUpdateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	defer func() {
		if err != nil {
			if err = RollbackRegistrationUpdateMutation(ctx, usesMutation); err != nil {
				panic(err)
			}
		}
	}()

	if err = FinalizeRegistrationUpdateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return eventspg.RegistrationUpdateResponse{
		EventRegistration: usesMutation.Transient.Current,
	}, nil
}

func InitiateRegistrationUpdateMutation(ctx context.Context, mutation *eventspg.RegistrationUpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "events.InitiateRegistrationUpdateMutation", mutation)
	defer logger.Exit(ctx, "events.InitiateRegistrationUpdateMutation", &err)

	result, err := SelectRegistrations(ctx, eventspg.RegistrationSelectArgument{Query: mutation.Foundation.Query})
	if err != nil {
		return err
	}

	if len(result.EventRegistrations) == 0 {
		return nil
	}

	updateRegistration := mutation.Foundation.Populate(result.EventRegistrations[0])

	mutation.Transient = eventspg.RegistrationTransient{
		Current: result.EventRegistrations[0],
		Future:  updateRegistration,
	}

	return nil
}

func FinalizeRegistrationUpdateMutation(ctx context.Context, mutation *eventspg.RegistrationUpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "events.FinalizeRegistrationUpdateMutation", mutation)
	defer logger.Exit(ctx, "events.FinalizeRegistrationUpdateMutation", &err)

	err = gormdb.Update(ctx, mutation.Transient.Future, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackRegistrationUpdateMutation(ctx context.Context, mutation *eventspg.RegistrationUpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "events.RollbackRegistrationUpdateMutation", mutation)
	defer logger.Exit(ctx, "events.RollbackRegistrationUpdateMutation", &err)

	if mutation.Transient.Current.Id == "" {
		return nil
	}

	return FinalizeRegistrationUpdateMutation(ctx, &eventspg.RegistrationUpdateMutation{
		Foundation: mutation.Foundation,
		Transient: eventspg.RegistrationTransient{
			Current: mutation.Transient.Current,
			Future:  mutation.Transient.Previous,
		},
	})
}
