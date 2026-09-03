package credentials

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	credentialspg "github.com/Johncode248/Rodzic-Chat/project/common/credentials"
)

func Delete(ctx context.Context, argument credentialspg.DeleteArgument) (res credentialspg.DeleteResponse, err error) {
	ctx = logger.Enter(ctx, "credentials.Delete", argument)
	defer logger.Exit(ctx, "credentials.Delete", &res, &err)

	var (
		usesMutation = &credentialspg.DeleteMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = ProcessDeleteMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return res, nil
}

func ProcessDeleteMutation(ctx context.Context, mutation *credentialspg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "credentials.ProcessDeleteMutation", mutation)
	defer logger.Exit(ctx, "credentials.ProcessDeleteMutation", &err)

	_, err = gormdb.Delete[credentialspg.Credential](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	return nil
}
