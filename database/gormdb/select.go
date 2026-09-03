package gormdb

import (
	"context"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
)

func Select[T any](ctx context.Context, query databasepg.Query) (result []T, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tx := makeCondition(ctx, query)

	if err = tx.Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func CustomSelect[T any](ctx context.Context, rawSQL string, args ...interface{}) (result []T, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tx := GormDB.WithContext(ctx)

	if err = tx.Raw(rawSQL, args...).Scan(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}
