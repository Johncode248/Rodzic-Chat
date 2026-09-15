package places

import (
	"context"
	"strings"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	placespg "github.com/Johncode248/Rodzic-Chat/project/common/places"
)

// List returns places (optionally filtered by category and/or a
// name/address substring) with each one's recommendation count and
// whether accountId itself is among the recommenders -- computed live via
// a join rather than a cached counter on Place, so it can never drift.
//
// sortBy "name" orders alphabetically; anything else (including "" and the
// default "recommendations") orders by recommendation count descending,
// which is the whole point of this feature -- the most-recommended places
// surface first.
func List(ctx context.Context, accountId string, category string, query string, sortBy string) (res []placespg.PlaceWithStats, err error) {
	ctx = logger.Enter(ctx, "places.List", accountId, category, query, sortBy)
	defer logger.Exit(ctx, "places.List", &res, &err)

	var conditions []string
	var args []interface{}

	if category != "" {
		conditions = append(conditions, "p.category = ?")
		args = append(args, category)
	}
	if query != "" {
		conditions = append(conditions, "(p.name ILIKE ? OR p.address ILIKE ?)")
		like := "%" + query + "%"
		args = append(args, like, like)
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	orderBy := "recommendation_count DESC, p.name ASC"
	if sortBy == "name" {
		orderBy = "p.name ASC"
	}

	sql := `
		SELECT p.*,
		       COUNT(r.id) AS recommendation_count,
		       COALESCE(BOOL_OR(r.account_id = ?), false) AS recommended_by_me
		FROM places p
		LEFT JOIN recommendations r ON r.place_id = p.id
		` + where + `
		GROUP BY p.id
		ORDER BY ` + orderBy

	allArgs := append([]interface{}{accountId}, args...)

	res, err = gormdb.CustomSelect[placespg.PlaceWithStats](ctx, sql, allArgs...)
	if err != nil {
		return nil, err
	}

	return res, nil
}
