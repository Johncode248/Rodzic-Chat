package absences

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	absencespg "github.com/Johncode248/Rodzic-Chat/project/common/absences"
)

func Delete(ctx context.Context, argument absencespg.DeleteArgument) (res absencespg.DeleteResponse, err error) {
	ctx = logger.Enter(ctx, "absences.Delete", argument)
	defer logger.Exit(ctx, "absences.Delete", &res, &err)

	var (
		usesMutation = &absencespg.DeleteMutation{
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
		res.Absences = append(res.Absences, value.Previous)
	}

	return res, nil
}

func InitiateDeleteMutation(ctx context.Context, mutation *absencespg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "absences.InitiateDeleteMutation", mutation)
	defer logger.Exit(ctx, "absences.InitiateDeleteMutation", &err)

	result, err := Select(ctx, absencespg.SelectArgument{
		Query: mutation.Foundation.Query,
	})
	if err != nil {
		return err
	}

	if len(result.Absences) == 0 {
		return nil
	}

	for _, absence := range result.Absences {
		mutation.Transient = append(mutation.Transient, absencespg.Transient{
			Current: absence,
		})
	}

	return nil
}

func FinalizeDeleteMutation(ctx context.Context, mutation *absencespg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "absences.FinalizeDeleteMutation", mutation)
	defer logger.Exit(ctx, "absences.FinalizeDeleteMutation", &err)

	_, err = gormdb.Delete[absencespg.Absence](ctx, mutation.Foundation.Query)
	if err != nil {
		return err
	}

	for index := range mutation.Transient {
		mutation.Transient[index] = mutation.Transient[index].Update()
	}

	return nil
}

func RollbackDeleteMutation(ctx context.Context, mutation *absencespg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "absences.RollbackDeleteMutation", mutation)
	defer logger.Exit(ctx, "absences.RollbackDeleteMutation", &err)

	return routines.Select(
		func() error {
			functions := []func() error{}
			for _, transient := range mutation.Transient {
				if transient.Previous.Id == "" {
					continue
				}

				transient := transient
				functions = append(functions, func() error {
					return FinalizeCreateMutation(ctx, &absencespg.CreateMutation{
						Transient: absencespg.Transient{
							Future: transient.Previous,
						},
					})
				})
			}
			return routines.Select(functions...)
		},
	)
}
