package children

import (
	"context"
	"errors"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
)

// ErrChildrenLimitReached means the organization's ChildrenLimit (its paid
// capacity tier, see project/common/organizations/pricing.go) has already
// been reached -- adding this child would exceed what the placowka pays
// for. Surfaced whether the child is being added by staff directly or via a
// parent's registration link (project/service/internal/registrationlinks),
// since both paths call Create.
var ErrChildrenLimitReached = errors.New("this organization has reached its paid children limit")

func Create(ctx context.Context, argument childrenpg.CreateArgument) (res childrenpg.CreateResponse, err error) {
	ctx = logger.Enter(ctx, "children.Create", argument)
	defer logger.Exit(ctx, "children.Create", &res, &err)

	var (
		usesMutation = &childrenpg.CreateMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = InitiateCreateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	if err = ValidateCreateMutation(ctx, usesMutation); err != nil {
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

	return childrenpg.CreateResponse{
		Child: usesMutation.Transient.Current,
	}, nil
}

func InitiateCreateMutation(ctx context.Context, mutation *childrenpg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "children.CreateInitiator", mutation)
	defer logger.Exit(ctx, "children.CreateInitiator", &err)

	future := mutation.Foundation.Populate()

	mutation.Transient.Future = future

	return nil
}

func ValidateCreateMutation(ctx context.Context, mutation *childrenpg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "children.ValidateCreateMutation", mutation)
	defer logger.Exit(ctx, "children.ValidateCreateMutation", &err)

	organizationId := mutation.Foundation.OrganizationId

	orgRows, err := gormdb.Select[organizationspg.Organization](ctx, databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: organizationId}}},
		},
	})
	if err != nil {
		return err
	}
	if len(orgRows) == 0 {
		return errors.New("organization not found")
	}

	// ChildrenLimit of 0 only happens for organizations created before this
	// field existed -- treated as unlimited rather than zero capacity, so
	// existing placowki aren't suddenly locked out.
	limit := orgRows[0].ChildrenLimit
	if limit == 0 {
		return nil
	}

	count, err := CountForOrganization(ctx, organizationId)
	if err != nil {
		return err
	}
	if count >= limit {
		return ErrChildrenLimitReached
	}

	return nil
}

func FinalizeCreateMutation(ctx context.Context, mutation *childrenpg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "children.FinalizeCreateMutation", mutation)
	defer logger.Exit(ctx, "children.FinalizeCreateMutation", &err)

	err = gormdb.Create(ctx, mutation.Transient.Future)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

func RollbackCreateMutation(ctx context.Context, mutation *childrenpg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "children.RollbackCreateMutation", mutation)
	defer logger.Exit(ctx, "children.RollbackCreateMutation", &err)

	return routines.Select(
		func() (err error) {
			return FinalizeDeleteMutation(ctx, &childrenpg.DeleteMutation{
				Foundation: childrenpg.DeleteArgument{
					Query: databasepg.Query{
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
					},
				},
			})
		},
	)
}
