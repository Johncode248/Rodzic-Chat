package sessions

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	sessionspg "github.com/Johncode248/Rodzic-Chat/project/common/sessions"
)

func Select(ctx context.Context, argument sessionspg.SelectArgument) (res sessionspg.SelectResponse, err error) {
	ctx = logger.Enter(ctx, "sessions.Select", argument)
	defer logger.Exit(ctx, "sessions.Select", &res, &err)

	var (
		usesMutation = &sessionspg.SelectMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = ProcessSelectMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return sessionspg.SelectResponse{
		Sessions: usesMutation.Transient,
	}, nil
}

func ProcessSelectMutation(ctx context.Context, mutation *sessionspg.SelectMutation) (err error) {
	ctx = logger.Enter(ctx, "sessions.ProcessSelectMutation", mutation)
	defer logger.Exit(ctx, "sessions.ProcessSelectMutation", &err)

	result, err := gormdb.Select[sessionspg.Session](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = result

	return nil
}
