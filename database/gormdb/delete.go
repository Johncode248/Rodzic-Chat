package gormdb

import (
	"context"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
)

func Delete[T any](ctx context.Context, query databasepg.Query) (result []T, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tx := makeCondition(ctx, query)

	if err := tx.Delete(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}
