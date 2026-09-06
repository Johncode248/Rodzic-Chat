package passwordresets

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	passwordresetspg "github.com/Johncode248/Rodzic-Chat/project/common/passwordresets"
)

func Create(ctx context.Context, argument passwordresetspg.CreateArgument) (res passwordresetspg.CreateResponse, err error) {
	ctx = logger.Enter(ctx, "passwordresets.Create", argument)
	defer logger.Exit(ctx, "passwordresets.Create", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	future := argument.Populate()

	if err = gormdb.Create(ctx, future); err != nil {
		return res, err
	}

	return passwordresetspg.CreateResponse{
		PasswordReset: future,
	}, nil
}
