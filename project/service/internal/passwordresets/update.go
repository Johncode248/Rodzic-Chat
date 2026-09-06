package passwordresets

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	passwordresetspg "github.com/Johncode248/Rodzic-Chat/project/common/passwordresets"
)

func Update(ctx context.Context, argument passwordresetspg.UpdateArgument) (res passwordresetspg.UpdateResponse, err error) {
	ctx = logger.Enter(ctx, "passwordresets.Update", argument)
	defer logger.Exit(ctx, "passwordresets.Update", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	current, err := Select(ctx, passwordresetspg.SelectArgument{Query: argument.Query})
	if err != nil {
		return res, err
	}

	if len(current.PasswordResets) == 0 {
		return res, nil
	}

	future := argument.Populate(current.PasswordResets[0])

	if err = gormdb.Update(ctx, future, argument.Query); err != nil {
		return res, err
	}

	return passwordresetspg.UpdateResponse{
		PasswordReset: future,
	}, nil
}
