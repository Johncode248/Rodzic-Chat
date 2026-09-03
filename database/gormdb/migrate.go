package gormdb

import (
	"context"
	"log"
	"time"

	"github.com/Johncode248/Rodzic-Chat/logger"
)

func Migrate(ctx context.Context, object interface{}) (err error) {
	ctx = logger.Enter(ctx, "gormdb.Migrate", object)
	defer logger.Exit(ctx, "gormdb.Migrate", &err)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err = GormDB.AutoMigrate(object)
	if err != nil {
		log.Fatal("Błąd migracji:", err)
	}

	return nil
}
