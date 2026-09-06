package accounts

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
)

// SelectUnscoped bypasses GORM's soft-delete scope, so it can see accounts
// with DeletedAt set -- used for the director-facing "removed parents"
// view (Roster's ?status=removed). Every other read path in this package
// stays scoped (excludes soft-deleted rows) via the normal Select.
func SelectUnscoped(ctx context.Context, argument accountspg.SelectArgument) (res accountspg.SelectResponse, err error) {
	ctx = logger.Enter(ctx, "accounts.SelectUnscoped", argument)
	defer logger.Exit(ctx, "accounts.SelectUnscoped", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	result, err := gormdb.SelectUnscoped[accountspg.Account](ctx, argument.Query)
	if err != nil {
		return res, err
	}

	return accountspg.SelectResponse{
		Accounts: result,
	}, nil
}
