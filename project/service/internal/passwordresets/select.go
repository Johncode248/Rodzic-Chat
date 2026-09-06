package passwordresets

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	passwordresetspg "github.com/Johncode248/Rodzic-Chat/project/common/passwordresets"
)

func Select(ctx context.Context, argument passwordresetspg.SelectArgument) (res passwordresetspg.SelectResponse, err error) {
	ctx = logger.Enter(ctx, "passwordresets.Select", argument)
	defer logger.Exit(ctx, "passwordresets.Select", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	result, err := gormdb.Select[passwordresetspg.PasswordReset](ctx, argument.Query)
	if err != nil {
		return res, err
	}

	return passwordresetspg.SelectResponse{
		PasswordResets: result,
	}, nil
}
