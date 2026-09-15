package places

import (
	"context"
	"errors"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	placespg "github.com/Johncode248/Rodzic-Chat/project/common/places"
)

var ErrPlaceNotFound = errors.New("place not found")

// GetOne is List's single-row equivalent, used to build a fresh response
// right after recommending a place (see recommend.go).
func GetOne(ctx context.Context, accountId string, placeId string) (res placespg.PlaceWithStats, err error) {
	ctx = logger.Enter(ctx, "places.GetOne", accountId, placeId)
	defer logger.Exit(ctx, "places.GetOne", &res, &err)

	const sql = `
		SELECT p.*,
		       COUNT(r.id) AS recommendation_count,
		       COALESCE(BOOL_OR(r.account_id = ?), false) AS recommended_by_me
		FROM places p
		LEFT JOIN recommendations r ON r.place_id = p.id
		WHERE p.id = ?
		GROUP BY p.id`

	rows, err := gormdb.CustomSelect[placespg.PlaceWithStats](ctx, sql, accountId, placeId)
	if err != nil {
		return res, err
	}
	if len(rows) == 0 {
		return res, ErrPlaceNotFound
	}

	return rows[0], nil
}
