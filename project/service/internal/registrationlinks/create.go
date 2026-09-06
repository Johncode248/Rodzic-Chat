package registrationlinks

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	registrationlinkspg "github.com/Johncode248/Rodzic-Chat/project/common/registrationlinks"
)

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
			if err = RollbackCreateMutation(ctx, usesMutation); err != nil {
				panic(err)
			}
		}
	}()

	if err = FinalizeCreateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return registrationlinkspg.CreateResponse{
		RegistrationLink: usesMutation.Transient.Current,
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
