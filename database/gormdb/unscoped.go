package gormdb

import (
	"context"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
)

// SelectUnscoped and DeleteUnscoped mirror Select/Delete but bypass GORM's
// soft-delete scope (Unscoped()), so they can see/remove rows a model's
// gorm.DeletedAt field would otherwise hide. Only cmd/reaper needs these
// today, for the 3-month hard-delete sweep of soft-deleted accounts.
func SelectUnscoped[T any](ctx context.Context, query databasepg.Query) (result []T, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tx := makeCondition(ctx, query).Unscoped()

	if err = tx.Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func DeleteUnscoped[T any](ctx context.Context, query databasepg.Query) (result []T, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tx := makeCondition(ctx, query).Unscoped()

	if err := tx.Delete(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}
