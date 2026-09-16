package groups

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	groupspg "github.com/Johncode248/Rodzic-Chat/project/common/groups"
)

func Update(ctx context.Context, argument groupspg.UpdateArgument) (res groupspg.UpdateResponse, err error) {
	ctx = logger.Enter(ctx, "groups.Update", argument)
	defer logger.Exit(ctx, "groups.Update", &res, &err)

	var (
		usesMutation = &groupspg.UpdateMutation{
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

	return groupspg.UpdateResponse{
		Group: usesMutation.Transient.Current,
	}, nil
}

func InitiateUpdateMutation(ctx context.Context, mutation *groupspg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "groups.InitiateUpdateMutation", mutation)
	defer logger.Exit(ctx, "groups.InitiateUpdateMutation", &err)

	result, err := Select(ctx, groupspg.SelectArgument{Query: mutation.Foundation.Query})
	if err != nil {
		return err
	}

	if len(result.Groups) == 0 {
		return nil
	}

	updateGroup := mutation.Foundation.Populate(result.Groups[0])

	mutation.Transient = groupspg.Transient{
		Current: result.Groups[0],
		Future:  updateGroup,
	}

	return nil
}

func FinalizeUpdateMutation(ctx context.Context, mutation *groupspg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "groups.FinalizeUpdateMutation", mutation)
	defer logger.Exit(ctx, "groups.FinalizeUpdateMutation", &err)

	err = gormdb.Update(ctx, mutation.Transient.Future, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackUpdateMutation(ctx context.Context, mutation *groupspg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "groups.RollbackUpdateMutation", mutation)
	defer logger.Exit(ctx, "groups.RollbackUpdateMutation", &err)

	return FinalizeUpdateMutation(ctx, &groupspg.UpdateMutation{
		Foundation: mutation.Foundation,
		Transient: groupspg.Transient{
			Current: mutation.Transient.Current,
			Future:  mutation.Transient.Previous,
		},
	})
}
