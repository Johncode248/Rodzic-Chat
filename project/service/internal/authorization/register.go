package authorization

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	credentialspg "github.com/Johncode248/Rodzic-Chat/project/common/credentials"
	featurespg "github.com/Johncode248/Rodzic-Chat/project/common/features"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/credentials"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/features"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/organizations"
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
			if rollbackErr := RollbackRegisterMutation(ctx, usesMutation); rollbackErr != nil {
				panic(rollbackErr)
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
		TokenType:    "Bearer",
		ExpiresIn:    int(credentials.AccessTokenTTL.Seconds()),
		ExpiresAt:    time.Now().Add(credentials.AccessTokenTTL).Unix(),
	}, nil
}

func ValidateRegisterMutation(ctx context.Context, mutation *authorizationpg.RegisterMutation) (err error) {
	ctx = logger.Enter(ctx, "authorization.ValidateRegisterMutation", mutation)
	defer logger.Exit(ctx, "authorization.ValidateRegisterMutation", &err)

	existingAccount, err := accounts.Select(ctx, accountspg.SelectArgument{
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

	if len(existingAccount.Accounts) > 0 {
		return errors.New("account already exists")
	}

	return nil
}

func FinalizeRegisterMutation(ctx context.Context, mutation *authorizationpg.RegisterMutation) (err error) {
	ctx = logger.Enter(ctx, "authorization.FinalizeRegisterMutation", mutation)
	defer logger.Exit(ctx, "authorization.FinalizeRegisterMutation", &err)

	// Registration always creates a brand-new organization first, since the
	// director account being created needs the new organization's id.
	organizationType, err := authorizationpg.MapOrganizationType(mutation.Foundation.OrganizationType)
	if err != nil {
		return err
	}

	organizationResult, err := organizations.Create(ctx, organizationspg.CreateArgument{
		Name:          mutation.Foundation.OrganizationName,
		Type:          organizationType,
		ChildrenLimit: mutation.Foundation.OrganizationChildrenLimit,
	})
	if err != nil {
		return err
	}

	mutation.Condition.Organization = organizationResult.Organization
	organizationId := organizationResult.Organization.Id

	// Seed one OrganizationFeature row per known feature so the director's
	// settings UI always has an explicit toggle to show, and so
	// features.IsEnabled's "no row" default-false path stays defensive-only.
	if err = seedOrganizationFeatures(ctx, organizationId); err != nil {
		return err
	}

	if err = routines.Select(
		func() error {
			accountId := uuid.NewString()
			return routines.Select(
				func() error {
					accountResult, err := accounts.Create(ctx, accountspg.CreateArgument{
						Id:             accountId,
						Name:           mutation.Foundation.Name,
						Email:          mutation.Foundation.Email,
						Phone:          mutation.Foundation.Phone,
						OrganizationId: organizationId,
						AccountType:    accountspg.AccountTypeDirector,
						Status:         accountspg.StatusEnabled,
					})
					if err != nil {
						return err
					}

					mutation.Condition.Account = accountResult.Account

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

	issued, err := credentials.IssueTokens(ctx, credentials.IssueTokensArgument{
		AccountId: mutation.Condition.Account.Id,
		Email:     mutation.Condition.Account.Email,
	})
	if err != nil {
		return err
	}

	mutation.Condition.AccessToken = issued.AccessToken
	mutation.Condition.RefreshToken = issued.RefreshToken

	return nil
}

// defaultFeatureEnabled decides the out-of-the-box state for a freshly
// created organization: everything on, including meal cancellations --
// that module is fully built (create/list/search/date-range) as of Faza 7,
// so it no longer needs to default off pending backend work.
func defaultFeatureEnabled(feature featurespg.Feature) bool {
	// Payments needs a completed Stripe Connect onboarding to do anything
	// useful (see project/common/organizationstripeaccounts) -- defaulting
	// it on would surface a "Płatności" module that silently can't charge
	// anyone yet, so it stays off until a director explicitly turns it on
	// (normally right after finishing onboarding).
	if feature == featurespg.FeaturePayments {
		return false
	}
	return true
}

// SeedOrganizationFeatures is the exported entry point used by profile creation.
func SeedOrganizationFeatures(ctx context.Context, organizationId string) error {
	return seedOrganizationFeatures(ctx, organizationId)
}

func seedOrganizationFeatures(ctx context.Context, organizationId string) error {
	all := featurespg.AllFeatures()
	functions := make([]func() error, 0, len(all))
	for _, feature := range all {
		feature := feature
		functions = append(functions, func() error {
			_, err := features.Create(ctx, featurespg.CreateArgument{
				OrganizationId: organizationId,
				Feature:        feature,
				Enabled:        defaultFeatureEnabled(feature),
			})
			return err
		})
	}

	return routines.Select(functions...)
}

func RollbackRegisterMutation(ctx context.Context, mutation *authorizationpg.RegisterMutation) (err error) {
	ctx = logger.Enter(ctx, "authorization.RollbackRegisterMutation", mutation)
	defer logger.Exit(ctx, "authorization.RollbackRegisterMutation", &err)

	functions := []func() error{
		func() (err error) {
			return accounts.FinalizeDeleteMutation(ctx, &accountspg.DeleteMutation{
				Foundation: accountspg.DeleteArgument{
					Query: databasepg.Query{
						Groups: []databasepg.Group{
							{
								Filters: []databasepg.Filter{
									{
										Key:      "id",
										Operator: "=",
										Value:    mutation.Condition.Account.Id,
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
										Key:      "user_id",
										Operator: "=",
										Value:    mutation.Condition.Account.Id,
									},
								},
							},
						},
					},
				},
			})
		},
	}

	if mutation.Condition.Organization.Id != "" {
		functions = append(functions, func() error {
			return organizations.FinalizeDeleteMutation(ctx, &organizationspg.DeleteMutation{
				Foundation: organizationspg.DeleteArgument{
					Query: databasepg.Query{
						Groups: []databasepg.Group{
							{
								Filters: []databasepg.Filter{
									{
										Key:      "id",
										Operator: "=",
										Value:    mutation.Condition.Organization.Id,
									},
								},
							},
						},
					},
				},
			})
		})
	}

	return routines.Select(functions...)
}
