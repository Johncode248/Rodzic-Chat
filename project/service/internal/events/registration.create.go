package events

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	eventspg "github.com/Johncode248/Rodzic-Chat/project/common/events"
)

func CreateRegistration(ctx context.Context, argument eventspg.RegistrationCreateArgument) (res eventspg.RegistrationCreateResponse, err error) {
	ctx = logger.Enter(ctx, "events.CreateRegistration", argument)
	defer logger.Exit(ctx, "events.CreateRegistration", &res, &err)

	var (
		usesMutation = &eventspg.RegistrationCreateMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = InitiateRegistrationCreateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	defer func() {
		if err != nil {
			if err = RollbackRegistrationCreateMutation(ctx, usesMutation); err != nil {
				panic(err)
			}
		}
	}()

	if err = FinalizeRegistrationCreateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return eventspg.RegistrationCreateResponse{
		EventRegistration: usesMutation.Transient.Current,
	}, nil
}

func InitiateRegistrationCreateMutation(ctx context.Context, mutation *eventspg.RegistrationCreateMutation) (err error) {
	ctx = logger.Enter(ctx, "events.InitiateRegistrationCreateMutation", mutation)
	defer logger.Exit(ctx, "events.InitiateRegistrationCreateMutation", &err)

	future := mutation.Foundation.Populate()

	mutation.Transient.Future = future

	return nil
}

func FinalizeRegistrationCreateMutation(ctx context.Context, mutation *eventspg.RegistrationCreateMutation) (err error) {
	ctx = logger.Enter(ctx, "events.FinalizeRegistrationCreateMutation", mutation)
	defer logger.Exit(ctx, "events.FinalizeRegistrationCreateMutation", &err)

	err = gormdb.Create(ctx, mutation.Transient.Future)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackRegistrationCreateMutation(ctx context.Context, mutation *eventspg.RegistrationCreateMutation) (err error) {
	ctx = logger.Enter(ctx, "events.RollbackRegistrationCreateMutation", mutation)
	defer logger.Exit(ctx, "events.RollbackRegistrationCreateMutation", &err)

	return routines.Select(
		func() (err error) {
			return FinalizeRegistrationDeleteMutation(ctx, &eventspg.RegistrationDeleteMutation{
				Foundation: eventspg.RegistrationDeleteArgument{
					Query: databasepg.Query{
						Groups: []databasepg.Group{
							{
								Filters: []databasepg.Filter{
									{
										Key:      "id",
										Operator: "=",
										Value:    mutation.Transient.Current.Id,
									},
								},
							},
						},
					},
				},
			})
		},
	)
}
