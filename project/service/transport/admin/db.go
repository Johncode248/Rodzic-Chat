package admingt

import (
	"context"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
)

func queryInto(r *http.Request, dest any, sql string, args ...any) error {
	return gormdb.GormDB.WithContext(r.Context()).Raw(sql, args...).Scan(dest).Error
}

func gormdbScan(ctx context.Context, dest any, sql string, args ...any) error {
	return gormdb.GormDB.WithContext(ctx).Raw(sql, args...).Scan(dest).Error
}
