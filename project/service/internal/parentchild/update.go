package parentchild

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
)

func Update(ctx context.Context, argument parentchildpg.UpdateArgument) (res parentchildpg.UpdateResponse, err error) {
	ctx = logger.Enter(ctx, "parentchild.Update", argument)
	defer logger.Exit(ctx, "parentchild.Update", &res, &err)

	var (
		usesMutation = &parentchildpg.UpdateMutation{
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

	return parentchildpg.UpdateResponse{
		ParentChild: usesMutation.Transient.Current,
	}, nil
}

func InitiateUpdateMutation(ctx context.Context, mutation *parentchildpg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "parentchild.InitiateUpdateMutation", mutation)
	defer logger.Exit(ctx, "parentchild.InitiateUpdateMutation", &err)

	result, err := Select(ctx, parentchildpg.SelectArgument{Query: mutation.Foundation.Query})
	if err != nil {
		return err
	}

	if len(result.ParentChildren) == 0 {
		return nil
	}

	updateParentChild := mutation.Foundation.Populate(result.ParentChildren[0])

	mutation.Transient = parentchildpg.Transient{
		Current: result.ParentChildren[0],
		Future:  updateParentChild,
	}

	return nil
}

func FinalizeUpdateMutation(ctx context.Context, mutation *parentchildpg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "parentchild.FinalizeUpdateMutation", mutation)
	defer logger.Exit(ctx, "parentchild.FinalizeUpdateMutation", &err)

	err = gormdb.Update(ctx, mutation.Transient.Future, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackUpdateMutation(ctx context.Context, mutation *parentchildpg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "parentchild.RollbackUpdateMutation", mutation)
	defer logger.Exit(ctx, "parentchild.RollbackUpdateMutation", &err)

	return FinalizeUpdateMutation(ctx, &parentchildpg.UpdateMutation{
		Foundation: mutation.Foundation,
		Transient: parentchildpg.Transient{
			Current: mutation.Transient.Current,
			Future:  mutation.Transient.Previous,
		},
	})
}
