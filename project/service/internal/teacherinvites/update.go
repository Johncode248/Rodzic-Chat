package teacherinvites

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	teacherinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/teacherinvites"
)

func Update(ctx context.Context, argument teacherinvitespg.UpdateArgument) (res teacherinvitespg.UpdateResponse, err error) {
	ctx = logger.Enter(ctx, "teacherinvites.Update", argument)
	defer logger.Exit(ctx, "teacherinvites.Update", &res, &err)

	var (
		usesMutation = &teacherinvitespg.UpdateMutation{
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

	return teacherinvitespg.UpdateResponse{
		TeacherInvite: usesMutation.Transient.Current,
	}, nil
}

func InitiateUpdateMutation(ctx context.Context, mutation *teacherinvitespg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "teacherinvites.InitiateUpdateMutation", mutation)
	defer logger.Exit(ctx, "teacherinvites.InitiateUpdateMutation", &err)

	result, err := Select(ctx, teacherinvitespg.SelectArgument{Query: mutation.Foundation.Query})
	if err != nil {
		return err
	}

	if len(result.TeacherInvites) == 0 {
		return nil
	}

	updateInvite := mutation.Foundation.Populate(result.TeacherInvites[0])

	mutation.Transient = teacherinvitespg.Transient{
		Current: result.TeacherInvites[0],
		Future:  updateInvite,
	}

	return nil
}

func FinalizeUpdateMutation(ctx context.Context, mutation *teacherinvitespg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "teacherinvites.FinalizeUpdateMutation", mutation)
	defer logger.Exit(ctx, "teacherinvites.FinalizeUpdateMutation", &err)

	if mutation.Transient.Current.Id == "" {
		return nil
	}

	err = gormdb.Update(ctx, mutation.Transient.Future, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackUpdateMutation(ctx context.Context, mutation *teacherinvitespg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "teacherinvites.RollbackUpdateMutation", mutation)
	defer logger.Exit(ctx, "teacherinvites.RollbackUpdateMutation", &err)

	if mutation.Transient.Current.Id == "" {
		return nil
	}

	return FinalizeUpdateMutation(ctx, &teacherinvitespg.UpdateMutation{
		Foundation: mutation.Foundation,
		Transient: teacherinvitespg.Transient{
			Current: mutation.Transient.Current,
			Future:  mutation.Transient.Previous,
		},
	})
}
