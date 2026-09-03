package credentials

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	credentialspg "github.com/Johncode248/Rodzic-Chat/project/common/credentials"
)

func Select(ctx context.Context, argument credentialspg.SelectArgument) (res credentialspg.SelectResponse, err error) {
	ctx = logger.Enter(ctx, "credentials.Select", argument)
	defer logger.Exit(ctx, "credentials.Select", &res, &err)

	var (
		usesMutation = &credentialspg.SelectMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = ProcessSelectMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return credentialspg.SelectResponse{
		Credentials: usesMutation.Transient,
	}, nil
}

func ProcessSelectMutation(ctx context.Context, mutation *credentialspg.SelectMutation) (err error) {
	ctx = logger.Enter(ctx, "credentials.ProcessSelectMutation", mutation)
	defer logger.Exit(ctx, "credentials.ProcessSelectMutation", &err)

	result, err := gormdb.Select[credentialspg.Credential](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = result

	return nil
}
