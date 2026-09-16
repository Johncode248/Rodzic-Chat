package children

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
)

func Update(ctx context.Context, argument childrenpg.UpdateArgument) (res childrenpg.UpdateResponse, err error) {
	ctx = logger.Enter(ctx, "children.Update", argument)
	defer logger.Exit(ctx, "children.Update", &res, &err)

	var (
		usesMutation = &childrenpg.UpdateMutation{
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
			if rollbackErr := RollbackUpdateMutation(ctx, usesMutation); rollbackErr != nil {
				panic(rollbackErr)
			}
		}
	}()

	if err = FinalizeUpdateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return childrenpg.UpdateResponse{
		Child: usesMutation.Transient.Current,
	}, nil
}

func InitiateUpdateMutation(ctx context.Context, mutation *childrenpg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "children.InitiateUpdateMutation", mutation)
	defer logger.Exit(ctx, "children.InitiateUpdateMutation", &err)

	result, err := Select(ctx, childrenpg.SelectArgument{Query: mutation.Foundation.Query})
	if err != nil {
		return err
	}

	if len(result.Children) == 0 {
		return nil
	}

	updateChild := mutation.Foundation.Populate(result.Children[0])

	mutation.Transient = childrenpg.Transient{
		Current: result.Children[0],
		Future:  updateChild,
	}

	return nil
}

func FinalizeUpdateMutation(ctx context.Context, mutation *childrenpg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "children.FinalizeUpdateMutation", mutation)
	defer logger.Exit(ctx, "children.FinalizeUpdateMutation", &err)

	err = gormdb.Update(ctx, mutation.Transient.Future, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackUpdateMutation(ctx context.Context, mutation *childrenpg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "children.RollbackUpdateMutation", mutation)
	defer logger.Exit(ctx, "children.RollbackUpdateMutation", &err)

	return FinalizeUpdateMutation(ctx, &childrenpg.UpdateMutation{
		Foundation: mutation.Foundation,
		Transient: childrenpg.Transient{
			Current: mutation.Transient.Current,
			Future:  mutation.Transient.Previous,
		},
	})
}
