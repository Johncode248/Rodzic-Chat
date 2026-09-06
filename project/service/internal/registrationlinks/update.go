package registrationlinks

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	registrationlinkspg "github.com/Johncode248/Rodzic-Chat/project/common/registrationlinks"
)

func Update(ctx context.Context, argument registrationlinkspg.UpdateArgument) (res registrationlinkspg.UpdateResponse, err error) {
	ctx = logger.Enter(ctx, "registrationlinks.Update", argument)
	defer logger.Exit(ctx, "registrationlinks.Update", &res, &err)

	var (
		usesMutation = &registrationlinkspg.UpdateMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = InitiateUpdateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	defer func() {
		if err != nil {
			if err = RollbackUpdateMutation(ctx, usesMutation); err != nil {
				panic(err)
			}
		}
	}()

	if err = FinalizeUpdateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return registrationlinkspg.UpdateResponse{
		RegistrationLink: usesMutation.Transient.Current,
	}, nil
}

func InitiateUpdateMutation(ctx context.Context, mutation *registrationlinkspg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "registrationlinks.InitiateUpdateMutation", mutation)
	defer logger.Exit(ctx, "registrationlinks.InitiateUpdateMutation", &err)

	result, err := Select(ctx, registrationlinkspg.SelectArgument{Query: mutation.Foundation.Query})
	if err != nil {
		return err
	}

	if len(result.RegistrationLinks) == 0 {
		return nil
	}

	updateLink := mutation.Foundation.Populate(result.RegistrationLinks[0])

	mutation.Transient = registrationlinkspg.Transient{
		Current: result.RegistrationLinks[0],
		Future:  updateLink,
	}

	return nil
}

func FinalizeUpdateMutation(ctx context.Context, mutation *registrationlinkspg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "registrationlinks.FinalizeUpdateMutation", mutation)
	defer logger.Exit(ctx, "registrationlinks.FinalizeUpdateMutation", &err)

	err = gormdb.Update(ctx, mutation.Transient.Future, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackUpdateMutation(ctx context.Context, mutation *registrationlinkspg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "registrationlinks.RollbackUpdateMutation", mutation)
	defer logger.Exit(ctx, "registrationlinks.RollbackUpdateMutation", &err)

	if mutation.Transient.Current.Id == "" {
		return nil
	}

	return FinalizeUpdateMutation(ctx, &registrationlinkspg.UpdateMutation{
		Foundation: mutation.Foundation,
		Transient: registrationlinkspg.Transient{
			Current: mutation.Transient.Current,
			Future:  mutation.Transient.Previous,
		},
	})
}
