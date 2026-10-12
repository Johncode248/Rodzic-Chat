package teacherinvites

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/pinhash"
	"github.com/Johncode248/Rodzic-Chat/library/random"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	teacherinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/teacherinvites"
)

// maxPinGenerationAttempts bounds retrying GeneratePin on the rare
// uniqueIndex collision on pin_hash.
const maxPinGenerationAttempts = 5

func Create(ctx context.Context, argument teacherinvitespg.CreateArgument) (res teacherinvitespg.CreateResponse, err error) {
	ctx = logger.Enter(ctx, "teacherinvites.Create", argument)
	defer logger.Exit(ctx, "teacherinvites.Create", &res, &err)

	var (
		usesMutation = &teacherinvitespg.CreateMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = InitiateCreateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	defer func() {
		if err != nil {
			if rollbackErr := RollbackCreateMutation(ctx, usesMutation); rollbackErr != nil {
				panic(rollbackErr)
			}
		}
	}()

	var plaintextPin string
	for attempt := 0; ; attempt++ {
		plaintextPin = random.GeneratePin()
		usesMutation.Transient.Future.PinHash = pinhash.HashPin(plaintextPin)

		err = FinalizeCreateMutation(ctx, usesMutation)
		if err == nil {
			break
		}
		if !pinhash.IsUniqueViolation(err) || attempt >= maxPinGenerationAttempts-1 {
			return res, err
		}
	}

	return teacherinvitespg.CreateResponse{
		TeacherInvite: usesMutation.Transient.Current,
		Pin:           plaintextPin,
	}, nil
}

func InitiateCreateMutation(ctx context.Context, mutation *teacherinvitespg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "teacherinvites.InitiateCreateMutation", mutation)
	defer logger.Exit(ctx, "teacherinvites.InitiateCreateMutation", &err)

	mutation.Transient.Future = mutation.Foundation.Populate()

	return nil
}

func FinalizeCreateMutation(ctx context.Context, mutation *teacherinvitespg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "teacherinvites.FinalizeCreateMutation", mutation)
	defer logger.Exit(ctx, "teacherinvites.FinalizeCreateMutation", &err)

	err = gormdb.Create(ctx, mutation.Transient.Future)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackCreateMutation(ctx context.Context, mutation *teacherinvitespg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "teacherinvites.RollbackCreateMutation", mutation)
	defer logger.Exit(ctx, "teacherinvites.RollbackCreateMutation", &err)

	return routines.Select(
		func() (err error) {
			return FinalizeDeleteMutation(ctx, &teacherinvitespg.DeleteMutation{
				Foundation: teacherinvitespg.DeleteArgument{
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
