package guardianclaims

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	guardianclaimspg "github.com/Johncode248/Rodzic-Chat/project/common/guardianclaims"
)

func Update(ctx context.Context, argument guardianclaimspg.UpdateArgument) (res guardianclaimspg.UpdateResponse, err error) {
	ctx = logger.Enter(ctx, "guardianclaims.Update", argument)
	defer logger.Exit(ctx, "guardianclaims.Update", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	current, err := Select(ctx, guardianclaimspg.SelectArgument{Query: argument.Query})
	if err != nil {
		return res, err
	}
	if len(current.GuardianClaims) == 0 {
		return res, nil
	}

	future := argument.Populate(current.GuardianClaims[0])

	if err = gormdb.Update(ctx, future, argument.Query); err != nil {
		return res, err
	}

	return guardianclaimspg.UpdateResponse{
		GuardianClaim: future,
	}, nil
}
