package groups

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	groupspg "github.com/Johncode248/Rodzic-Chat/project/common/groups"
)

func Delete(ctx context.Context, argument groupspg.DeleteArgument) (res groupspg.DeleteResponse, err error) {
	ctx = logger.Enter(ctx, "groups.Delete", argument)
	defer logger.Exit(ctx, "groups.Delete", &res, &err)

	var (
		usesMutation = &groupspg.DeleteMutation{
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
		res.Groups = append(res.Groups, value.Previous)
	}

	return res, nil
}

func InitiateDeleteMutation(ctx context.Context, mutation *groupspg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "groups.InitiateDeleteMutation", mutation)
	defer logger.Exit(ctx, "groups.InitiateDeleteMutation", &err)

	result, err := Select(ctx, groupspg.SelectArgument{
		Query: mutation.Foundation.Query,
	})
	if err != nil {
		return err
	}

	if len(result.Groups) == 0 {
		return nil
	}

	for _, group := range result.Groups {
		mutation.Transient = append(mutation.Transient, groupspg.Transient{
			Current: group,
		})
	}

	return nil
}

func FinalizeDeleteMutation(ctx context.Context, mutation *groupspg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "groups.FinalizeDeleteMutation", mutation)
	defer logger.Exit(ctx, "groups.FinalizeDeleteMutation", &err)

	_, err = gormdb.Delete[groupspg.Group](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	for index := range mutation.Transient {
		mutation.Transient[index] = mutation.Transient[index].Update()
	}

	return nil
}

func RollbackDeleteMutation(ctx context.Context, mutation *groupspg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "groups.RollbackDeleteMutation", mutation)
	defer logger.Exit(ctx, "groups.RollbackDeleteMutation", &err)

	return routines.Select(
		func() error {
			functions := []func() error{}
			for _, transient := range mutation.Transient {
				if transient.Previous.Id == "" {
					continue
				}

				functions = append(functions, func() error {
					return FinalizeCreateMutation(ctx, &groupspg.CreateMutation{
						Transient: groupspg.Transient{
							Future: transient.Previous,
						},
					})
				})
			}
			return routines.Select(functions...)
		},
	)
}
