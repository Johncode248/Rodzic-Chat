package registrationlinks

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	registrationlinkspg "github.com/Johncode248/Rodzic-Chat/project/common/registrationlinks"
)

func Select(ctx context.Context, argument registrationlinkspg.SelectArgument) (res registrationlinkspg.SelectResponse, err error) {
	ctx = logger.Enter(ctx, "registrationlinks.Select", argument)
	defer logger.Exit(ctx, "registrationlinks.Select", &res, &err)

	var (
		usesMutation = &registrationlinkspg.SelectMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = ProcessSelectMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return registrationlinkspg.SelectResponse{
		RegistrationLinks: usesMutation.Transient,
	}, nil
}

func ProcessSelectMutation(ctx context.Context, mutation *registrationlinkspg.SelectMutation) (err error) {
	ctx = logger.Enter(ctx, "registrationlinks.ProcessSelectMutation", mutation)
	defer logger.Exit(ctx, "registrationlinks.ProcessSelectMutation", &err)

	result, err := gormdb.Select[registrationlinkspg.RegistrationLink](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = result

	return nil
}
