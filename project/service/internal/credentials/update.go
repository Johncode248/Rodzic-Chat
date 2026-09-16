package credentials

import (
	"context"
	"errors"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	credentialspg "github.com/Johncode248/Rodzic-Chat/project/common/credentials"
)

var ErrMissingUserId = errors.New("missing user id")

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

	// An account that only ever logged in via Google/Apple has no
	// Credential row at all -- if the caller is trying to SET a password
	// (Password != nil), that must create one, not silently no-op. Without
	// this, PATCH /private/credentials returned 200 for such an account
	// with nothing written, and the caller still couldn't log in with a
	// password afterward.
	if argument.Password != nil {
		existing, selectErr := Select(ctx, credentialspg.SelectArgument{Query: argument.Query})
		if selectErr != nil {
			return res, selectErr
		}
		if len(existing.Credentials) == 0 {
			if argument.UserId == "" {
				return res, ErrMissingUserId
			}
			createResult, createErr := Create(ctx, credentialspg.CreateArgument{
				UserId:   argument.UserId,
				Password: *argument.Password,
			})
			if createErr != nil {
				return res, createErr
			}
			return credentialspg.UpdateResponse{Credential: createResult.Credential}, nil
		}
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
