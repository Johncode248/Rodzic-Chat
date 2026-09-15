package children

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
)

func Select(ctx context.Context, argument childrenpg.SelectArgument) (res childrenpg.SelectResponse, err error) {
	ctx = logger.Enter(ctx, "children.Select", argument)
	defer logger.Exit(ctx, "children.Select", &res, &err)

	var (
		usesMutation = &childrenpg.SelectMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = ProcessSelectMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return childrenpg.SelectResponse{
		Children: usesMutation.Transient,
	}, nil
}

func ProcessSelectMutation(ctx context.Context, mutation *childrenpg.SelectMutation) (err error) {
	ctx = logger.Enter(ctx, "children.ProcessSelectMutation", mutation)
	defer logger.Exit(ctx, "children.ProcessSelectMutation", &err)

	result, err := gormdb.Select[childrenpg.Child](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	for i := range result {
		result[i].HasPesel = result[i].EncryptedPesel != ""
	}

	mutation.Transient = result

	return nil
}
