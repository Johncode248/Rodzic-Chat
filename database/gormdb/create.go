package gormdb

import (
	"context"
	"time"

	"github.com/Johncode248/Rodzic-Chat/logger"
)

func Create(ctx context.Context, object interface{}) (err error) {
	ctx = logger.Enter(ctx, "gormdb.Create", object)
	defer logger.Exit(ctx, "gormdb.Create", &err)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err = GormDB.WithContext(ctx).Create(object).Error; err != nil {
		return err
	}

	return nil
}
