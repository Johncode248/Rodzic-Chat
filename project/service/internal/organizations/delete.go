package organizations

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
)

func Delete(ctx context.Context, argument organizationspg.DeleteArgument) (res organizationspg.DeleteResponse, err error) {
	ctx = logger.Enter(ctx, "organizations.Delete", argument)
	defer logger.Exit(ctx, "organizations.Delete", &res, &err)

	var (
		usesMutation = &organizationspg.DeleteMutation{
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
			if rollbackErr := RollbackDeleteMutation(ctx, usesMutation); rollbackErr != nil {
				panic(rollbackErr)
			}
		}
	}()

	if err = FinalizeDeleteMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	for _, value := range usesMutation.Transient {
		res.Organizations = append(res.Organizations, value.Previous)
	}

	return res, nil
}

func InitiateDeleteMutation(ctx context.Context, mutation *organizationspg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "organizations.InitiateDeleteMutation", mutation)
	defer logger.Exit(ctx, "organizations.InitiateDeleteMutation", &err)

	result, err := Select(ctx, organizationspg.SelectArgument{
		Query: mutation.Foundation.Query,
	})
	if err != nil {
		return err
	}

	if len(result.Organizations) == 0 {
		return nil
	}

	for _, organization := range result.Organizations {
		mutation.Transient = append(mutation.Transient, organizationspg.Transient{
			Current: organization,
		})
	}

	return nil
}

func FinalizeDeleteMutation(ctx context.Context, mutation *organizationspg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "organizations.FinalizeDeleteMutation", mutation)
	defer logger.Exit(ctx, "organizations.FinalizeDeleteMutation", &err)

	_, err = gormdb.Delete[organizationspg.Organization](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	for index := range mutation.Transient {
		mutation.Transient[index] = mutation.Transient[index].Update()
	}

	return nil
}

func RollbackDeleteMutation(ctx context.Context, mutation *organizationspg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "organizations.RollbackDeleteMutation", mutation)
	defer logger.Exit(ctx, "organizations.RollbackDeleteMutation", &err)

	return routines.Select(
		func() error {
			functions := []func() error{}
			for _, transient := range mutation.Transient {
				if transient.Previous.Id == "" {
					continue
				}

				functions = append(functions, func() error {
					return FinalizeCreateMutation(ctx, &organizationspg.CreateMutation{
						Transient: organizationspg.Transient{
							Future: transient.Previous,
						},
					})
				})
			}
			return routines.Select(functions...)
		},
	)
}
