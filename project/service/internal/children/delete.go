package children

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
)

func Delete(ctx context.Context, argument childrenpg.DeleteArgument) (res childrenpg.DeleteResponse, err error) {
	ctx = logger.Enter(ctx, "children.Delete", argument)
	defer logger.Exit(ctx, "children.Delete", &res, &err)

	var (
		usesMutation = &childrenpg.DeleteMutation{
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
		res.Children = append(res.Children, value.Previous)
	}

	return res, nil
}

func InitiateDeleteMutation(ctx context.Context, mutation *childrenpg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "children.InitiateDeleteMutation", mutation)
	defer logger.Exit(ctx, "children.InitiateDeleteMutation", &err)

	result, err := Select(ctx, childrenpg.SelectArgument{
		Query: mutation.Foundation.Query,
	})
	if err != nil {
		return err
	}

	if len(result.Children) == 0 {
		return nil
	}

	for _, child := range result.Children {
		mutation.Transient = append(mutation.Transient, childrenpg.Transient{
			Current: child,
		})
	}

	return nil
}

func FinalizeDeleteMutation(ctx context.Context, mutation *childrenpg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "children.FinalizeDeleteMutation", mutation)
	defer logger.Exit(ctx, "children.FinalizeDeleteMutation", &err)

	_, err = gormdb.Delete[childrenpg.Child](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	for index := range mutation.Transient {
		mutation.Transient[index] = mutation.Transient[index].Update()
	}

	return nil
}

func RollbackDeleteMutation(ctx context.Context, mutation *childrenpg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "children.RollbackDeleteMutation", mutation)
	defer logger.Exit(ctx, "children.RollbackDeleteMutation", &err)

	return routines.Select(
		func() error {
			functions := []func() error{}
			for _, transient := range mutation.Transient {
				if transient.Previous.Id == "" {
					continue
				}

				functions = append(functions, func() error {
					return FinalizeCreateMutation(ctx, &childrenpg.CreateMutation{
						Transient: childrenpg.Transient{
							Future: transient.Previous,
						},
					})
				})
			}
			return routines.Select(functions...)
		},
	)
}
