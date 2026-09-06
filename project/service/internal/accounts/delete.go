package accounts

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	credentialspg "github.com/Johncode248/Rodzic-Chat/project/common/credentials"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/credentials"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/sessions"
)

func Delete(ctx context.Context, argument accountspg.DeleteArgument) (res accountspg.DeleteResponse, err error) {
	ctx = logger.Enter(ctx, "accounts.Delete", argument)
	defer logger.Exit(ctx, "accounts.Delete", &res, &err)

	var (
		usesMutation = &accountspg.DeleteMutation{
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
			if err = RollbackDeleteMutation(ctx, usesMutation); err != nil {
				panic(err)
			}
		}
	}()

	if err = FinalizeDeleteMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	for _, value := range usesMutation.Transient {
		res.Accounts = append(res.Accounts, value.Previous)
	}

	return res, nil
}

func InitiateDeleteMutation(ctx context.Context, mutation *accountspg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "accounts.InitiateDeleteMutation", mutation)
	defer logger.Exit(ctx, "accounts.InitiateDeleteMutation", &err)

	result, err := Select(ctx, accountspg.SelectArgument{
		Query: mutation.Foundation.Query,
	})
	if err != nil {
		return err
	}

	if len(result.Accounts) == 0 {
		return nil
	}

	for _, account := range result.Accounts {
		mutation.Transient = append(mutation.Transient, accountspg.Transient{
			Current: account,
		})
	}

	functions := []func() error{}
	for _, transient := range mutation.Transient {
		{
			dependence := &credentialspg.DeleteMutation{
				Foundation: credentialspg.DeleteArgument{
					Query: databasepg.Query{
						Groups: []databasepg.Group{
							{
								Filters: []databasepg.Filter{
									{
										Key:      "user_id",
										Operator: "=",
										Value:    transient.Current.Id,
									},
								},
							},
						},
					},
				},
			}

			mutation.Condition.DeleteCredentialMutation = append(mutation.Condition.DeleteCredentialMutation, dependence)
		}
	}

	return routines.Select(functions...)
}

func FinalizeDeleteMutation(ctx context.Context, mutation *accountspg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "accounts.FinalizeDeleteMutation", mutation)
	defer logger.Exit(ctx, "accounts.FinalizeDeleteMutation", &err)

	_, err = gormdb.Delete[accountspg.Account](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	for index := range mutation.Transient {
		mutation.Transient[index] = mutation.Transient[index].Update()
	}

	functions := []func() error{}
	for _, condition := range mutation.Condition.DeleteCredentialMutation {
		functions = append(functions, func() error {
			return credentials.ProcessDeleteMutation(ctx, condition)
		})
	}

	// Deleting an account (self-service or otherwise) is a security event:
	// every device logged in on it should be forced to re-authenticate, the
	// same as a password reset. ParentChild rows are deliberately left
	// alone here -- a child may still need the historical link during the
	// 3-month soft-delete grace window -- only cmd/reaper removes those.
	//
	// transient.Previous (not .Current) holds the deleted account here: the
	// .Update() loop above already rotated Current -> Previous and zeroed
	// Current to Future's zero value, since Delete never sets Future.
	for _, transient := range mutation.Transient {
		accountId := transient.Previous.Id
		functions = append(functions, func() error {
			return sessions.RevokeAllForAccount(ctx, accountId)
		})
	}

	return routines.Select(functions...)
}

func RollbackDeleteMutation(ctx context.Context, mutation *accountspg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "accounts.RollbackDeleteMutation", mutation)
	defer logger.Exit(ctx, "accounts.RollbackDeleteMutation", &err)

	return routines.Select(
		func() error {
			functions := []func() error{}
			for _, transient := range mutation.Transient {
				if transient.Previous.Id == "" {
					continue
				}

				functions = append(functions, func() error {
					return FinalizeCreateMutation(ctx, &accountspg.CreateMutation{
						Transient: accountspg.Transient{
							Future: transient.Previous,
						},
					})
				})
			}
			return routines.Select(functions...)
		},
	)
}
