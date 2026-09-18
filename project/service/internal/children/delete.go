package children

import (
	"context"
	"time"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
)

func Delete(ctx context.Context, argument childrenpg.DeleteArgument) (res childrenpg.DeleteResponse, err error) {
	ctx = logger.Enter(ctx, "children.Delete", argument)
	defer logger.Exit(ctx, "children.Delete", &res, &err)

	var (
		usesMutation = &childrenpg.DeleteMutation{
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
			if rollbackErr := RollbackDeleteMutation(ctx, usesMutation); rollbackErr != nil {
				panic(rollbackErr)
			}
		}
	}()

	if err = FinalizeDeleteMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	for _, value := range usesMutation.Transient {
		res.Children = append(res.Children, value.Previous)
	}

	return res, nil
}

func InitiateDeleteMutation(ctx context.Context, mutation *childrenpg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "children.InitiateDeleteMutation", mutation)
	defer logger.Exit(ctx, "children.InitiateDeleteMutation", &err)

	result, err := Select(ctx, childrenpg.SelectArgument{
		Query: mutation.Foundation.Query,
	})
	if err != nil {
		return err
	}

	if len(result.Children) == 0 {
		return nil
	}

	for _, child := range result.Children {
		mutation.Transient = append(mutation.Transient, childrenpg.Transient{
			Current: child,
		})
	}

	return nil
}

// FinalizeDeleteMutation never hard-deletes the Child row: a placówka
// "removing" a child from its roster must not take that child away from
// the parent, who keeps seeing it (and its chat/event/absence history)
// afterwards -- see project/common/children/child.go's doc comment on
// OrganizationId. Unenrolling just clears OrganizationId/GroupId on every
// matched child; the row and every ParentChild link to it are untouched.
func FinalizeDeleteMutation(ctx context.Context, mutation *childrenpg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "children.FinalizeDeleteMutation", mutation)
	defer logger.Exit(ctx, "children.FinalizeDeleteMutation", &err)

	for index := range mutation.Transient {
		unenrolled := mutation.Transient[index].Current
		unenrolled.OrganizationId = ""
		unenrolled.GroupId = ""
		unenrolled.UpdateTime = time.Now()

		if err = gormdb.Update(ctx, &unenrolled, byIdQuery(unenrolled.Id)); err != nil {
			return err
		}

		mutation.Transient[index].Future = unenrolled
	}

	for index := range mutation.Transient {
		mutation.Transient[index] = mutation.Transient[index].Update()
	}

	return nil
}

func RollbackDeleteMutation(ctx context.Context, mutation *childrenpg.DeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "children.RollbackDeleteMutation", mutation)
	defer logger.Exit(ctx, "children.RollbackDeleteMutation", &err)

	return routines.Select(
		func() error {
			functions := []func() error{}
			for _, transient := range mutation.Transient {
				if transient.Previous.Id == "" {
					continue
				}

				restored := transient.Previous
				functions = append(functions, func() error {
					return gormdb.Update(ctx, &restored, byIdQuery(restored.Id))
				})
			}
			return routines.Select(functions...)
		},
	)
}
