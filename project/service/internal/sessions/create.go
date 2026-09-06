package sessions

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	sessionspg "github.com/Johncode248/Rodzic-Chat/project/common/sessions"
)

func Create(ctx context.Context, argument sessionspg.CreateArgument) (res sessionspg.CreateResponse, err error) {
	ctx = logger.Enter(ctx, "sessions.Create", argument)
	defer logger.Exit(ctx, "sessions.Create", &res, &err)

	var (
		usesMutation = &sessionspg.CreateMutation{
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

	return sessionspg.CreateResponse{
		Session: usesMutation.Transient.Current,
	}, nil
}

func InitiateCreateMutation(ctx context.Context, mutation *sessionspg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "sessions.InitiateCreateMutation", mutation)
	defer logger.Exit(ctx, "sessions.InitiateCreateMutation", &err)

	mutation.Transient.Future = mutation.Foundation.Populate()

	return nil
}

func FinalizeCreateMutation(ctx context.Context, mutation *sessionspg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "sessions.FinalizeCreateMutation", mutation)
	defer logger.Exit(ctx, "sessions.FinalizeCreateMutation", &err)

	err = gormdb.Create(ctx, mutation.Transient.Future)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackCreateMutation(ctx context.Context, mutation *sessionspg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "sessions.RollbackCreateMutation", mutation)
	defer logger.Exit(ctx, "sessions.RollbackCreateMutation", &err)

	if mutation.Transient.Current.Id == "" {
		return nil
	}

	return routines.Select(
		func() error {
			_, err := gormdb.Delete[sessionspg.Session](ctx, databasepg.Query{
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
			})
			return err
		},
	)
}
