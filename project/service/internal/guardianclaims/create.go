package guardianclaims

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	guardianclaimspg "github.com/Johncode248/Rodzic-Chat/project/common/guardianclaims"
)

func Create(ctx context.Context, argument guardianclaimspg.CreateArgument) (res guardianclaimspg.CreateResponse, err error) {
	ctx = logger.Enter(ctx, "guardianclaims.Create", argument)
	defer logger.Exit(ctx, "guardianclaims.Create", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	future := argument.Populate()

	if err = gormdb.Create(ctx, future); err != nil {
		return res, err
	}

	return guardianclaimspg.CreateResponse{
		GuardianClaim: future,
	}, nil
}
