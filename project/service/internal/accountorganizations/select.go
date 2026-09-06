package accountorganizations

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountorganizationspg "github.com/Johncode248/Rodzic-Chat/project/common/accountorganizations"
)

func Select(ctx context.Context, argument accountorganizationspg.SelectArgument) (res accountorganizationspg.SelectResponse, err error) {
	ctx = logger.Enter(ctx, "accountorganizations.Select", argument)
	defer logger.Exit(ctx, "accountorganizations.Select", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	result, err := gormdb.Select[accountorganizationspg.AccountOrganization](ctx, argument.Query)
	if err != nil {
		return res, err
	}

	return accountorganizationspg.SelectResponse{
		AccountOrganizations: result,
	}, nil
}
