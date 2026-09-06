package parentchild

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
)

func Delete(ctx context.Context, argument parentchildpg.DeleteArgument) (res parentchildpg.DeleteResponse, err error) {
	ctx = logger.Enter(ctx, "parentchild.Delete", argument)
	defer logger.Exit(ctx, "parentchild.Delete", &res, &err)

	var (
		usesMutation = &parentchildpg.DeleteMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = InitiateDeleteMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	defer func() {
		if err != nil {
			if err = RollbackDeleteMutation(ctx, usesMutation); err != nil {
				panic(err)
			}
		}
	}()

	if err = FinalizeDeleteMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	for _, value := range usesMutation.Transient {
		res.ParentChildren = append(res.ParentChildren, value.Previous)
	}

	return res, nil
}

func InitiateDeleteMutation(ctx context.Context, mutation *parentchildpg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "parentchild.InitiateDeleteMutation", mutation)
	defer logger.Exit(ctx, "parentchild.InitiateDeleteMutation", &err)

	result, err := Select(ctx, parentchildpg.SelectArgument{
		Query: mutation.Foundation.Query,
	})
	if err != nil {
		return err
	}

	if len(result.ParentChildren) == 0 {
		return nil
	}

	for _, parentChild := range result.ParentChildren {
		mutation.Transient = append(mutation.Transient, parentchildpg.Transient{
			Current: parentChild,
		})
	}

	return nil
}

func FinalizeDeleteMutation(ctx context.Context, mutation *parentchildpg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "parentchild.FinalizeDeleteMutation", mutation)
	defer logger.Exit(ctx, "parentchild.FinalizeDeleteMutation", &err)

	_, err = gormdb.Delete[parentchildpg.ParentChild](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	for index := range mutation.Transient {
		mutation.Transient[index] = mutation.Transient[index].Update()
	}

	return nil
}

func RollbackDeleteMutation(ctx context.Context, mutation *parentchildpg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "parentchild.RollbackDeleteMutation", mutation)
	defer logger.Exit(ctx, "parentchild.RollbackDeleteMutation", &err)

	return routines.Select(
		func() error {
			functions := []func() error{}
			for _, transient := range mutation.Transient {
				if transient.Previous.Id == "" {
					continue
				}

				functions = append(functions, func() error {
					return FinalizeCreateMutation(ctx, &parentchildpg.CreateMutation{
						Transient: parentchildpg.Transient{
							Future: transient.Previous,
						},
					})
				})
			}
			return routines.Select(functions...)
		},
	)
}
