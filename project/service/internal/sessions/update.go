package sessions

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	sessionspg "github.com/Johncode248/Rodzic-Chat/project/common/sessions"
)

func Update(ctx context.Context, argument sessionspg.UpdateArgument) (res sessionspg.UpdateResponse, err error) {
	ctx = logger.Enter(ctx, "sessions.Update", argument)
	defer logger.Exit(ctx, "sessions.Update", &res, &err)

	var (
		usesMutation = &sessionspg.UpdateMutation{
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
			if err = RollbackUpdateMutation(ctx, usesMutation); err != nil {
				panic(err)
			}
		}
	}()

	if err = FinalizeUpdateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return sessionspg.UpdateResponse{
		Session: usesMutation.Transient.Current,
	}, nil
}

func InitiateUpdateMutation(ctx context.Context, mutation *sessionspg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "sessions.InitiateUpdateMutation", mutation)
	defer logger.Exit(ctx, "sessions.InitiateUpdateMutation", &err)

	result, err := Select(ctx, sessionspg.SelectArgument{Query: mutation.Foundation.Query})
	if err != nil {
		return err
	}

	if len(result.Sessions) == 0 {
		return nil
	}

	mutation.Transient = sessionspg.Transient{
		Current: result.Sessions[0],
		Future:  mutation.Foundation.Populate(result.Sessions[0]),
	}

	return nil
}

func FinalizeUpdateMutation(ctx context.Context, mutation *sessionspg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "sessions.FinalizeUpdateMutation", mutation)
	defer logger.Exit(ctx, "sessions.FinalizeUpdateMutation", &err)

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

func RollbackUpdateMutation(ctx context.Context, mutation *sessionspg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "sessions.RollbackUpdateMutation", mutation)
	defer logger.Exit(ctx, "sessions.RollbackUpdateMutation", &err)

	if mutation.Transient.Previous.Id == "" {
		return nil
	}

	return FinalizeUpdateMutation(ctx, &sessionspg.UpdateMutation{
		Foundation: mutation.Foundation,
		Transient: sessionspg.Transient{
			Current: mutation.Transient.Current,
			Future:  mutation.Transient.Previous,
		},
	})
}
