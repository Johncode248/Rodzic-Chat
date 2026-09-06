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

	// Select("*") is required here: GORM's struct-based Updates() silently
	// skips zero-value fields (false, "", 0, zero time) since it can't tell
	// "explicitly cleared" from "not set" on a plain struct. Every internal
	// Update flow already populates `model` as a complete row (Populate()
	// copies the fetched current row, then applies only the requested
	// partial changes), so writing every column back is safe -- without
	// this, e.g. setting a bool field to false, or a string field to "",
	// would be silently ignored and the old value would stick.
	if err := tx.Select("*").Updates(model).Error; err != nil {
		return err
	}

	return nil
}
