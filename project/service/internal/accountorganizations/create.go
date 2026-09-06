package accountorganizations

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountorganizationspg "github.com/Johncode248/Rodzic-Chat/project/common/accountorganizations"
)

func Create(ctx context.Context, argument accountorganizationspg.CreateArgument) (res accountorganizationspg.CreateResponse, err error) {
	ctx = logger.Enter(ctx, "accountorganizations.Create", argument)
	defer logger.Exit(ctx, "accountorganizations.Create", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	future := argument.Populate()

	if err = gormdb.Create(ctx, future); err != nil {
		return res, err
	}

	return accountorganizationspg.CreateResponse{
		AccountOrganization: future,
	}, nil
}
