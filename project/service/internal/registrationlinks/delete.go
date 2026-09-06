package registrationlinks

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	registrationlinkspg "github.com/Johncode248/Rodzic-Chat/project/common/registrationlinks"
)

func Delete(ctx context.Context, argument registrationlinkspg.DeleteArgument) (res registrationlinkspg.DeleteResponse, err error) {
	ctx = logger.Enter(ctx, "registrationlinks.Delete", argument)
	defer logger.Exit(ctx, "registrationlinks.Delete", &res, &err)

	var (
		usesMutation = &registrationlinkspg.DeleteMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = InitiateDeleteMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	defer func() {
		if err != nil {
			if err = RollbackDeleteMutation(ctx, usesMutation); err != nil {
				panic(err)
			}
		}
	}()

	if err = FinalizeDeleteMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	for _, value := range usesMutation.Transient {
		res.RegistrationLinks = append(res.RegistrationLinks, value.Previous)
	}

	return res, nil
}

func InitiateDeleteMutation(ctx context.Context, mutation *registrationlinkspg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "registrationlinks.InitiateDeleteMutation", mutation)
	defer logger.Exit(ctx, "registrationlinks.InitiateDeleteMutation", &err)

	result, err := Select(ctx, registrationlinkspg.SelectArgument{
		Query: mutation.Foundation.Query,
	})
	if err != nil {
		return err
	}

	if len(result.RegistrationLinks) == 0 {
		return nil
	}

	for _, link := range result.RegistrationLinks {
		mutation.Transient = append(mutation.Transient, registrationlinkspg.Transient{
			Current: link,
		})
	}

	return nil
}

func FinalizeDeleteMutation(ctx context.Context, mutation *registrationlinkspg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "registrationlinks.FinalizeDeleteMutation", mutation)
	defer logger.Exit(ctx, "registrationlinks.FinalizeDeleteMutation", &err)

	_, err = gormdb.Delete[registrationlinkspg.RegistrationLink](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	for index := range mutation.Transient {
		mutation.Transient[index] = mutation.Transient[index].Update()
	}

	return nil
}

func RollbackDeleteMutation(ctx context.Context, mutation *registrationlinkspg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "registrationlinks.RollbackDeleteMutation", mutation)
	defer logger.Exit(ctx, "registrationlinks.RollbackDeleteMutation", &err)

	return routines.Select(
		func() error {
			functions := []func() error{}
			for _, transient := range mutation.Transient {
				if transient.Previous.Id == "" {
					continue
				}

				transient := transient
				functions = append(functions, func() error {
					return FinalizeCreateMutation(ctx, &registrationlinkspg.CreateMutation{
						Transient: registrationlinkspg.Transient{
							Future: transient.Previous,
						},
					})
				})
			}
			return routines.Select(functions...)
		},
	)
}
