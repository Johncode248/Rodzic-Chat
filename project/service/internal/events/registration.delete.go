package events

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	eventspg "github.com/Johncode248/Rodzic-Chat/project/common/events"
)

func DeleteRegistration(ctx context.Context, argument eventspg.RegistrationDeleteArgument) (res eventspg.RegistrationDeleteResponse, err error) {
	ctx = logger.Enter(ctx, "events.DeleteRegistration", argument)
	defer logger.Exit(ctx, "events.DeleteRegistration", &res, &err)

	var (
		usesMutation = &eventspg.RegistrationDeleteMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = InitiateRegistrationDeleteMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	defer func() {
		if err != nil {
			if err = RollbackRegistrationDeleteMutation(ctx, usesMutation); err != nil {
				panic(err)
			}
		}
	}()

	if err = FinalizeRegistrationDeleteMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	for _, value := range usesMutation.Transient {
		res.EventRegistrations = append(res.EventRegistrations, value.Previous)
	}

	return res, nil
}

func InitiateRegistrationDeleteMutation(ctx context.Context, mutation *eventspg.RegistrationDeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "events.InitiateRegistrationDeleteMutation", mutation)
	defer logger.Exit(ctx, "events.InitiateRegistrationDeleteMutation", &err)

	result, err := SelectRegistrations(ctx, eventspg.RegistrationSelectArgument{
		Query: mutation.Foundation.Query,
	})
	if err != nil {
		return err
	}

	if len(result.EventRegistrations) == 0 {
		return nil
	}

	for _, registration := range result.EventRegistrations {
		mutation.Transient = append(mutation.Transient, eventspg.RegistrationTransient{
			Current: registration,
		})
	}

	return nil
}

func FinalizeRegistrationDeleteMutation(ctx context.Context, mutation *eventspg.RegistrationDeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "events.FinalizeRegistrationDeleteMutation", mutation)
	defer logger.Exit(ctx, "events.FinalizeRegistrationDeleteMutation", &err)

	_, err = gormdb.Delete[eventspg.EventRegistration](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	for index := range mutation.Transient {
		mutation.Transient[index] = mutation.Transient[index].Update()
	}

	return nil
}

func RollbackRegistrationDeleteMutation(ctx context.Context, mutation *eventspg.RegistrationDeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "events.RollbackRegistrationDeleteMutation", mutation)
	defer logger.Exit(ctx, "events.RollbackRegistrationDeleteMutation", &err)

	return routines.Select(
		func() error {
			functions := []func() error{}
			for _, transient := range mutation.Transient {
				if transient.Previous.Id == "" {
					continue
				}

				transient := transient
				functions = append(functions, func() error {
					return FinalizeRegistrationCreateMutation(ctx, &eventspg.RegistrationCreateMutation{
						Transient: eventspg.RegistrationTransient{
							Future: transient.Previous,
						},
					})
				})
			}
			return routines.Select(functions...)
		},
	)
}
