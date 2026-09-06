package guardianclaims

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	guardianclaimspg "github.com/Johncode248/Rodzic-Chat/project/common/guardianclaims"
)

func Select(ctx context.Context, argument guardianclaimspg.SelectArgument) (res guardianclaimspg.SelectResponse, err error) {
	ctx = logger.Enter(ctx, "guardianclaims.Select", argument)
	defer logger.Exit(ctx, "guardianclaims.Select", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	result, err := gormdb.Select[guardianclaimspg.GuardianClaim](ctx, argument.Query)
	if err != nil {
		return res, err
	}

	return guardianclaimspg.SelectResponse{
		GuardianClaims: result,
	}, nil
}
