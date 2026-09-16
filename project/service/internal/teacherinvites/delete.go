package teacherinvites

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	teacherinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/teacherinvites"
)

func Delete(ctx context.Context, argument teacherinvitespg.DeleteArgument) (res teacherinvitespg.DeleteResponse, err error) {
	ctx = logger.Enter(ctx, "teacherinvites.Delete", argument)
	defer logger.Exit(ctx, "teacherinvites.Delete", &res, &err)

	var (
		usesMutation = &teacherinvitespg.DeleteMutation{
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
		res.TeacherInvites = append(res.TeacherInvites, value.Previous)
	}

	return res, nil
}

func InitiateDeleteMutation(ctx context.Context, mutation *teacherinvitespg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "teacherinvites.InitiateDeleteMutation", mutation)
	defer logger.Exit(ctx, "teacherinvites.InitiateDeleteMutation", &err)

	result, err := Select(ctx, teacherinvitespg.SelectArgument{
		Query: mutation.Foundation.Query,
	})
	if err != nil {
		return err
	}

	if len(result.TeacherInvites) == 0 {
		return nil
	}

	for _, invite := range result.TeacherInvites {
		mutation.Transient = append(mutation.Transient, teacherinvitespg.Transient{
			Current: invite,
		})
	}

	return nil
}

func FinalizeDeleteMutation(ctx context.Context, mutation *teacherinvitespg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "teacherinvites.FinalizeDeleteMutation", mutation)
	defer logger.Exit(ctx, "teacherinvites.FinalizeDeleteMutation", &err)

	_, err = gormdb.Delete[teacherinvitespg.TeacherInvite](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	for index := range mutation.Transient {
		mutation.Transient[index] = mutation.Transient[index].Update()
	}

	return nil
}

func RollbackDeleteMutation(ctx context.Context, mutation *teacherinvitespg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "teacherinvites.RollbackDeleteMutation", mutation)
	defer logger.Exit(ctx, "teacherinvites.RollbackDeleteMutation", &err)

	return routines.Select(
		func() error {
			functions := []func() error{}
			for _, transient := range mutation.Transient {
				if transient.Previous.Id == "" {
					continue
				}

				transient := transient
				functions = append(functions, func() error {
					return FinalizeCreateMutation(ctx, &teacherinvitespg.CreateMutation{
						Transient: teacherinvitespg.Transient{
							Future: transient.Previous,
						},
					})
				})
			}
			return routines.Select(functions...)
		},
	)
}
