package features

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	featurespg "github.com/Johncode248/Rodzic-Chat/project/common/features"
)

func Delete(ctx context.Context, argument featurespg.DeleteArgument) (res featurespg.DeleteResponse, err error) {
	ctx = logger.Enter(ctx, "features.Delete", argument)
	defer logger.Exit(ctx, "features.Delete", &res, &err)

	var (
		usesMutation = &featurespg.DeleteMutation{
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
		res.OrganizationFeatures = append(res.OrganizationFeatures, value.Previous)
	}

	return res, nil
}

func InitiateDeleteMutation(ctx context.Context, mutation *featurespg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "features.InitiateDeleteMutation", mutation)
	defer logger.Exit(ctx, "features.InitiateDeleteMutation", &err)

	result, err := Select(ctx, featurespg.SelectArgument{
		Query: mutation.Foundation.Query,
	})
	if err != nil {
		return err
	}

	if len(result.OrganizationFeatures) == 0 {
		return nil
	}

	for _, organizationFeature := range result.OrganizationFeatures {
		mutation.Transient = append(mutation.Transient, featurespg.Transient{
			Current: organizationFeature,
		})
	}

	return nil
}

func FinalizeDeleteMutation(ctx context.Context, mutation *featurespg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "features.FinalizeDeleteMutation", mutation)
	defer logger.Exit(ctx, "features.FinalizeDeleteMutation", &err)

	_, err = gormdb.Delete[featurespg.OrganizationFeature](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	for index := range mutation.Transient {
		mutation.Transient[index] = mutation.Transient[index].Update()
	}

	return nil
}

func RollbackDeleteMutation(ctx context.Context, mutation *featurespg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "features.RollbackDeleteMutation", mutation)
	defer logger.Exit(ctx, "features.RollbackDeleteMutation", &err)

	return routines.Select(
		func() error {
			functions := []func() error{}
			for _, transient := range mutation.Transient {
				if transient.Previous.Id == "" {
					continue
				}

				transient := transient
				functions = append(functions, func() error {
					return FinalizeCreateMutation(ctx, &featurespg.CreateMutation{
						Transient: featurespg.Transient{
							Future: transient.Previous,
						},
					})
				})
			}
			return routines.Select(functions...)
		},
	)
}
