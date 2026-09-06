package teacherinvites

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	teacherinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/teacherinvites"
)

func Create(ctx context.Context, argument teacherinvitespg.CreateArgument) (res teacherinvitespg.CreateResponse, err error) {
	ctx = logger.Enter(ctx, "teacherinvites.Create", argument)
	defer logger.Exit(ctx, "teacherinvites.Create", &res, &err)

	var (
		usesMutation = &teacherinvitespg.CreateMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = InitiateCreateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	defer func() {
		if err != nil {
			if err = RollbackCreateMutation(ctx, usesMutation); err != nil {
				panic(err)
			}
		}
	}()

	if err = FinalizeCreateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return teacherinvitespg.CreateResponse{
		TeacherInvite: usesMutation.Transient.Current,
	}, nil
}

func InitiateCreateMutation(ctx context.Context, mutation *teacherinvitespg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "teacherinvites.InitiateCreateMutation", mutation)
	defer logger.Exit(ctx, "teacherinvites.InitiateCreateMutation", &err)

	mutation.Transient.Future = mutation.Foundation.Populate()

	return nil
}

func FinalizeCreateMutation(ctx context.Context, mutation *teacherinvitespg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "teacherinvites.FinalizeCreateMutation", mutation)
	defer logger.Exit(ctx, "teacherinvites.FinalizeCreateMutation", &err)

	err = gormdb.Create(ctx, mutation.Transient.Future)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackCreateMutation(ctx context.Context, mutation *teacherinvitespg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "teacherinvites.RollbackCreateMutation", mutation)
	defer logger.Exit(ctx, "teacherinvites.RollbackCreateMutation", &err)

	return routines.Select(
		func() (err error) {
			return FinalizeDeleteMutation(ctx, &teacherinvitespg.DeleteMutation{
				Foundation: teacherinvitespg.DeleteArgument{
					Query: databasepg.Query{
						Groups: []databasepg.Group{
							{
								Filters: []databasepg.Filter{
									{
										Key:      "id",
										Operator: "=",
										Value:    mutation.Transient.Current.Id,
									},
								},
							},
						},
					},
				},
			})
		},
	)
}
