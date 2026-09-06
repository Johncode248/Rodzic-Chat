package gormdb

import (
	"context"
	"fmt"
	"strings"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"gorm.io/gorm"
)

func makeCondition(ctx context.Context, query databasepg.Query) *gorm.DB {

	tx := GormDB.WithContext(ctx)

	var orConditions []string
	var orValues []interface{}

	for _, group := range query.Groups {
		var andConditions []string
		var andValues []interface{}

		for _, f := range group.Filters {
			// "IS NOT"/"IS" with a nil Value means a literal NULL check
			// (e.g. {Key: "deleted_at", Operator: "IS NOT", Value: nil} ->
			// "deleted_at IS NOT NULL") -- these have no bound parameter,
			// unlike every other operator here.
			if (f.Operator == "IS NOT" || f.Operator == "IS") && f.Value == nil {
				andConditions = append(andConditions, fmt.Sprintf("%s %s NULL", f.Key, f.Operator))
				continue
			}

			cond := fmt.Sprintf("%s %s ?", f.Key, f.Operator)
			andConditions = append(andConditions, cond)
			andValues = append(andValues, f.Value)
		}

		if len(andConditions) > 0 {
			orConditions = append(orConditions, "("+strings.Join(andConditions, " AND ")+")")
			orValues = append(orValues, andValues...)
		}
	}

	if len(orConditions) > 0 {
		where := strings.Join(orConditions, " OR ")
		tx = tx.Where(where, orValues...)
	}

	if query.Limit > 0 {
		tx = tx.Limit(query.Limit)
	}
	if query.OrderBy.Key != "" && query.OrderBy.Direction != databasepg.DirectionUnknown {
		tx = tx.Order(query.OrderBy.ToString())
	}

	if query.Offset > 0 {
		tx = tx.Offset(query.Offset)
	}

	return tx
}
