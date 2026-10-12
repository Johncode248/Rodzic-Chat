package registrationlinks

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/pinhash"
	"github.com/Johncode248/Rodzic-Chat/library/random"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	registrationlinkspg "github.com/Johncode248/Rodzic-Chat/project/common/registrationlinks"
)

// maxPinGenerationAttempts bounds retrying GeneratePin on the rare
// uniqueIndex collision on pin_hash -- an 8-digit PIN's keyspace (1e8) is
// small enough that a collision is a real (if unlikely) possibility, unlike
// GenerateToken's 190-bit space.
const maxPinGenerationAttempts = 5

func Create(ctx context.Context, argument registrationlinkspg.CreateArgument) (res registrationlinkspg.CreateResponse, err error) {
	ctx = logger.Enter(ctx, "registrationlinks.Create", argument)
	defer logger.Exit(ctx, "registrationlinks.Create", &res, &err)

	var (
		usesMutation = &registrationlinkspg.CreateMutation{
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
		// Collision on pin_hash -- retry with a freshly generated PIN.
		// FinalizeCreateMutation already advanced Transient on failure? No
		// -- Update() only ever runs on success, so Future still holds the
		// same (unsaved) row; just overwrite PinHash and try again.
	}

	return registrationlinkspg.CreateResponse{
		RegistrationLink: usesMutation.Transient.Current,
		Pin:              plaintextPin,
	}, nil
}

func InitiateCreateMutation(ctx context.Context, mutation *registrationlinkspg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "registrationlinks.InitiateCreateMutation", mutation)
	defer logger.Exit(ctx, "registrationlinks.InitiateCreateMutation", &err)

	future := mutation.Foundation.Populate()

	mutation.Transient.Future = future

	return nil
}

func FinalizeCreateMutation(ctx context.Context, mutation *registrationlinkspg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "registrationlinks.FinalizeCreateMutation", mutation)
	defer logger.Exit(ctx, "registrationlinks.FinalizeCreateMutation", &err)

	err = gormdb.Create(ctx, mutation.Transient.Future)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackCreateMutation(ctx context.Context, mutation *registrationlinkspg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "registrationlinks.RollbackCreateMutation", mutation)
	defer logger.Exit(ctx, "registrationlinks.RollbackCreateMutation", &err)

	return routines.Select(
		func() (err error) {
			return FinalizeDeleteMutation(ctx, &registrationlinkspg.DeleteMutation{
				Foundation: registrationlinkspg.DeleteArgument{
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
