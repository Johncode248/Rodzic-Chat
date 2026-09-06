package accountorganizations

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountorganizationspg "github.com/Johncode248/Rodzic-Chat/project/common/accountorganizations"
)

func Delete(ctx context.Context, argument accountorganizationspg.DeleteArgument) (res accountorganizationspg.DeleteResponse, err error) {
	ctx = logger.Enter(ctx, "accountorganizations.Delete", argument)
	defer logger.Exit(ctx, "accountorganizations.Delete", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	deleted, err := gormdb.Delete[accountorganizationspg.AccountOrganization](ctx, argument.Query)
	if err != nil {
		return res, err
	}

	return accountorganizationspg.DeleteResponse{
		AccountOrganizations: deleted,
	}, nil
}
