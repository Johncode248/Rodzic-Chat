package places

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	placespg "github.com/Johncode248/Rodzic-Chat/project/common/places"
)

// Unrecommend retracts accountId's own recommendation from placeId, if
// any. The underlying Place row is left in place either way -- other
// parents' recommendations (or the place's mere existence) don't depend on
// any one recommender.
func Unrecommend(ctx context.Context, accountId string, placeId string) (err error) {
	ctx = logger.Enter(ctx, "places.Unrecommend", accountId, placeId)
	defer logger.Exit(ctx, "places.Unrecommend", &err)

	_, err = gormdb.Delete[placespg.Recommendation](ctx, byPlaceAndAccount(placeId, accountId))
	return err
}
