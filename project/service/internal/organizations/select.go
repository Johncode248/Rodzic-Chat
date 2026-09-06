package organizations

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
)

func Select(ctx context.Context, argument organizationspg.SelectArgument) (res organizationspg.SelectResponse, err error) {
	ctx = logger.Enter(ctx, "organizations.Select", argument)
	defer logger.Exit(ctx, "organizations.Select", &res, &err)

	var (
		usesMutation = &organizationspg.SelectMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = ProcessSelectMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return organizationspg.SelectResponse{
		Organizations: usesMutation.Transient,
	}, nil
}

func ProcessSelectMutation(ctx context.Context, mutation *organizationspg.SelectMutation) (err error) {
	ctx = logger.Enter(ctx, "organizations.ProcessSelectMutation", mutation)
	defer logger.Exit(ctx, "organizations.ProcessSelectMutation", &err)

	result, err := gormdb.Select[organizationspg.Organization](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	mutation.Transient = result

	return nil
}
