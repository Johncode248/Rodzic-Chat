package authorization

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	credentialspg "github.com/Johncode248/Rodzic-Chat/project/common/credentials"
	userspg "github.com/Johncode248/Rodzic-Chat/project/common/users"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/credentials"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/users"
	"github.com/google/uuid"
)

func Register(ctx context.Context, argument authorizationpg.RegisterArgument) (res authorizationpg.RegisterReply, err error) {
	ctx = logger.Enter(ctx, "authorization.Register", argument)
	defer logger.Exit(ctx, "authorization.Register", &res, &err)

	var (
		usesMutation = &authorizationpg.RegisterMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = ValidateRegisterMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	defer func() {
		if err != nil {
			if err = RollbackRegisterMutation(ctx, usesMutation); err != nil {
				panic(err)
			}
		}
	}()

	if err = FinalizeRegisterMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	//go func() {
	//	if err := SendRegisterEmail(ctx, argument.Email); err != nil {
	//		log.Fatalln(err.Error())
	//	}
	//}()

	return authorizationpg.RegisterReply{
		AccessToken:  usesMutation.Condition.AccessToken,
		RefreshToken: usesMutation.Condition.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Hour * 23).Unix(),
	}, nil
}

func ValidateRegisterMutation(ctx context.Context, mutation *authorizationpg.RegisterMutation) (err error) {
	ctx = logger.Enter(ctx, "authorization.ValidateRegisterMutation", mutation)
	defer logger.Exit(ctx, "authorization.ValidateRegisterMutation", &err)

	existingAccount, err := users.Select(ctx, userspg.SelectArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{
							Key:      "email",
							Operator: "=",
							Value:    mutation.Foundation.Email,
						},
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	if len(existingAccount.Users) > 0 {
		return errors.New("account already exists")
	}

	return nil
}

func FinalizeRegisterMutation(ctx context.Context, mutation *authorizationpg.RegisterMutation) (err error) {
	ctx = logger.Enter(ctx, "authorization.FinalizeRegisterMutation", mutation)
	defer logger.Exit(ctx, "authorization.FinalizeRegisterMutation", &err)

	if err = routines.Select(
		func() error {
			accountId := uuid.NewString()
			return routines.Select(
				func() error {
					accountResult, err := users.Create(ctx, userspg.CreateArgument{
						Id:     accountId,
						Name:   mutation.Foundation.Name,
						Email:  mutation.Foundation.Email,
						Phone:  mutation.Foundation.Phone,
						Status: userspg.StatusEnabled,
					})
					if err != nil {
						return err
					}

					mutation.Condition.User = accountResult.User

					return nil
				},
				func() error {
					if mutation.Foundation.Password == "" {
						return nil
					} else {
						credentialResult, err := credentials.Create(ctx, credentialspg.CreateArgument{
							UserId:   accountId,
							Password: mutation.Foundation.Password,
						})
						if err != nil {
							return err
						}

						mutation.Condition.Credential = credentialResult.Credential

						return nil
					}
				},
			)
		},
	); err != nil {
		return err
	}

	return routines.Select(
		func() error {
			accessToken, err := credentials.GenerateAccessToken(ctx, mutation.Condition.User.Id, mutation.Condition.User.Email)
			if err != nil {
				return err
			}

			mutation.Condition.AccessToken = accessToken

			return nil
		},
		func() error {
			refreshToken, err := credentials.GenerateRefreshToken(ctx, mutation.Condition.User.Id, mutation.Condition.User.Email)
			if err != nil {
				return err
			}

			mutation.Condition.RefreshToken = refreshToken

			return nil
		},
	)
}

func RollbackRegisterMutation(ctx context.Context, mutation *authorizationpg.RegisterMutation) (err error) {
	ctx = logger.Enter(ctx, "authorization.RollbackRegisterMutation", mutation)
	defer logger.Exit(ctx, "authorization.RollbackRegisterMutation", &err)

	return routines.Select(
		func() (err error) {
			return users.FinalizeDeleteMutation(ctx, &userspg.DeleteMutation{
				Foundation: userspg.DeleteArgument{
					Query: databasepg.Query{
						Groups: []databasepg.Group{
							{
								Filters: []databasepg.Filter{
									{
										Key:      "id",
										Operator: "=",
										Value:    mutation.Condition.User.Id,
									},
								},
							},
						},
					},
				},
			})
		},
		func() error {
			return credentials.ProcessDeleteMutation(ctx, &credentialspg.DeleteMutation{
				Foundation: credentialspg.DeleteArgument{
					Query: databasepg.Query{
						Groups: []databasepg.Group{
							{
								Filters: []databasepg.Filter{
									{
										Key:      "userId",
										Operator: "=",
										Value:    mutation.Condition.User.Id,
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
