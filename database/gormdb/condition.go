package gormdb

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"gorm.io/gorm"
)

// validColumnKey matches a bare identifier -- every Filter.Key in this
// codebase is a hardcoded literal set by trusted server code (verified
// repo-wide), never attacker-controlled, so this can never actually fire
// today. It exists as defense-in-depth: Key/Operator below are
// string-concatenated directly into the WHERE clause (only Value is a
// bound parameter), which would become a direct SQL-injection primitive
// -- arbitrary column name, or via Operator an arbitrary SQL fragment --
// the moment anything wires a client-supplied Query through here (several
// unused SelectRequest DTOs in project/common/*/select.argument.request.go
// already expose Filter.Key/Operator as JSON fields, ready to be
// connected to a handler by a future change). Failing loudly here is the
// point: a legitimate literal key/operator always matches, so this never
// changes behavior for any current caller.
var validColumnKey = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// validOperators is the fixed, exhaustive set of operators actually used
// across every Filter{} literal in this codebase (verified repo-wide).
var validOperators = map[string]bool{
	"=": true, "!=": true, "<": true, "<=": true, ">": true, ">=": true,
	"IN": true, "NOT IN": true, "LIKE": true, "ILIKE": true, "IS": true, "IS NOT": true,
}

func makeCondition(ctx context.Context, query databasepg.Query) *gorm.DB {

	tx := GormDB.WithContext(ctx)

	var orConditions []string
	var orValues []interface{}

	for _, group := range query.Groups {
		var andConditions []string
		var andValues []interface{}

		for _, f := range group.Filters {
			if !validColumnKey.MatchString(f.Key) || !validOperators[f.Operator] {
				panic(fmt.Sprintf("gormdb: refusing unsafe filter key=%q operator=%q", f.Key, f.Operator))
			}

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
		// Same defense-in-depth reasoning as the Filter.Key/Operator check
		// above -- Order.ToString() string-concatenates Key directly into
		// the ORDER BY clause with no parameterization possible for an
		// identifier position.
		if !validColumnKey.MatchString(query.OrderBy.Key) {
			panic(fmt.Sprintf("gormdb: refusing unsafe order key=%q", query.OrderBy.Key))
		}
		tx = tx.Order(query.OrderBy.ToString())
	}

	if query.Offset > 0 {
		tx = tx.Offset(query.Offset)
	}

	return tx
}
