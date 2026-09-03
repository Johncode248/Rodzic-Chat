package credentials

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	credentialspg "github.com/Johncode248/Rodzic-Chat/project/common/credentials"
)

func Update(ctx context.Context, argument credentialspg.UpdateArgument) (res credentialspg.UpdateResponse, err error) {
	ctx = logger.Enter(ctx, "credentials.Update", argument)
	defer logger.Exit(ctx, "credentials.Update", &res, &err)

	var (
		usesMutation = &credentialspg.UpdateMutation{
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

	return credentialspg.UpdateResponse{
		Credential: usesMutation.Transient.Current,
	}, nil
}

func InitiateUpdateMutation(ctx context.Context, mutation *credentialspg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "credentials.InitiateUpdateMutation", mutation)
	defer logger.Exit(ctx, "credentials.InitiateUpdateMutation", &err)

	result, err := Select(ctx, credentialspg.SelectArgument{Query: mutation.Foundation.Query})
	if err != nil {
		return err
	}

	if len(result.Credentials) == 0 {
		return nil
	}

	updateCredential := mutation.Foundation.Populate(result.Credentials[0])

	mutation.Transient = credentialspg.Transient{
		Current: result.Credentials[0],
		Future:  updateCredential,
	}

	return nil
}

func FinalizeUpdateMutation(ctx context.Context, mutation *credentialspg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "credentials.FinalizeUpdateMutation", mutation)
	defer logger.Exit(ctx, "credentials.FinalizeUpdateMutation", &err)

	err = gormdb.Update(ctx, mutation.Transient.Future, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackUpdateMutation(ctx context.Context, mutation *credentialspg.UpdateMutation) (err error) {
	ctx = logger.Enter(ctx, "credentials.RollbackUpdateMutation", mutation)
	defer logger.Exit(ctx, "credentials.RollbackUpdateMutation", &err)

	return FinalizeUpdateMutation(ctx, &credentialspg.UpdateMutation{
		Foundation: mutation.Foundation,
		Transient: credentialspg.Transient{
			Current: mutation.Transient.Current,
			Future:  mutation.Transient.Previous,
		},
	})
}
