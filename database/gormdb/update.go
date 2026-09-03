package gormdb

import (
	"context"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
)

func Update(ctx context.Context, model interface{}, query databasepg.Query) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tx := makeCondition(ctx, query)

	if err := tx.Updates(model).Error; err != nil {
		return err
	}

	return nil
}
