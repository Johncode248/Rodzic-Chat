package events

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	eventspg "github.com/Johncode248/Rodzic-Chat/project/common/events"
)

func CreateEvent(ctx context.Context, argument eventspg.EventCreateArgument) (res eventspg.EventCreateResponse, err error) {
	ctx = logger.Enter(ctx, "events.CreateEvent", argument)
	defer logger.Exit(ctx, "events.CreateEvent", &res, &err)

	var (
		usesMutation = &eventspg.EventCreateMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = InitiateEventCreateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	defer func() {
		if err != nil {
			if err = RollbackEventCreateMutation(ctx, usesMutation); err != nil {
				panic(err)
			}
		}
	}()

	if err = FinalizeEventCreateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return eventspg.EventCreateResponse{
		Event: usesMutation.Transient.Current,
	}, nil
}

func InitiateEventCreateMutation(ctx context.Context, mutation *eventspg.EventCreateMutation) (err error) {
	ctx = logger.Enter(ctx, "events.InitiateEventCreateMutation", mutation)
	defer logger.Exit(ctx, "events.InitiateEventCreateMutation", &err)

	future := mutation.Foundation.Populate()

	mutation.Transient.Future = future

	return nil
}

func FinalizeEventCreateMutation(ctx context.Context, mutation *eventspg.EventCreateMutation) (err error) {
	ctx = logger.Enter(ctx, "events.FinalizeEventCreateMutation", mutation)
	defer logger.Exit(ctx, "events.FinalizeEventCreateMutation", &err)

	err = gormdb.Create(ctx, mutation.Transient.Future)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackEventCreateMutation(ctx context.Context, mutation *eventspg.EventCreateMutation) (err error) {
	ctx = logger.Enter(ctx, "events.RollbackEventCreateMutation", mutation)
	defer logger.Exit(ctx, "events.RollbackEventCreateMutation", &err)

	return routines.Select(
		func() (err error) {
			return FinalizeEventDeleteMutation(ctx, &eventspg.EventDeleteMutation{
				Foundation: eventspg.EventDeleteArgument{
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
