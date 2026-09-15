package places

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	parentsubscriptionspg "github.com/Johncode248/Rodzic-Chat/project/common/parentsubscriptions"
	placespg "github.com/Johncode248/Rodzic-Chat/project/common/places"
	"github.com/google/uuid"
)

// ErrRecommendationLimitReached is returned when accountId has already recommended
// parentsubscriptionspg.MaxRecommendationsPerParent distinct places -- see Recommend below. A
// repeat recommend of a place already on their list is never blocked by this (Recommend upserts
// rather than inserting in that case), only genuinely new ones.
var ErrRecommendationLimitReached = errors.New("recommendation limit reached")

func byAccountIdRecommendations(accountId string) databasepg.Query {
	return databasepg.Query{Groups: []databasepg.Group{
		{Filters: []databasepg.Filter{{Key: "account_id", Operator: "=", Value: accountId}}},
	}}
}

func byExternalRef(source string, ref string) databasepg.Query {
	return databasepg.Query{Limit: 1, Groups: []databasepg.Group{
		{Filters: []databasepg.Filter{
			{Key: "source", Operator: "=", Value: source},
			{Key: "external_ref", Operator: "=", Value: ref},
		}},
	}}
}

func byNameAndAddress(name string, address string) databasepg.Query {
	return databasepg.Query{Limit: 1, Groups: []databasepg.Group{
		{Filters: []databasepg.Filter{
			{Key: "name", Operator: "=", Value: name},
			{Key: "address", Operator: "=", Value: address},
			{Key: "source", Operator: "=", Value: placespg.SourceManual},
		}},
	}}
}

func byPlaceAndAccount(placeId string, accountId string) databasepg.Query {
	return databasepg.Query{Limit: 1, Groups: []databasepg.Group{
		{Filters: []databasepg.Filter{
			{Key: "place_id", Operator: "=", Value: placeId},
			{Key: "account_id", Operator: "=", Value: accountId},
		}},
	}}
}

// resolvePlace finds the existing Place a RecommendRequest refers to, or
// creates one. Deduplication: an explicit PlaceId always wins; an OSM pick
// is de-duplicated by its stable ExternalRef (so every parent who picks
// the same OSM result stacks recommendations on one row instead of
// spawning duplicates); a manually-typed place is de-duplicated by an
// exact name+address match, a reasonable heuristic given OSM has no
// listing to key off of.
func resolvePlace(ctx context.Context, accountId string, req placespg.RecommendRequest) (place placespg.Place, err error) {
	if req.PlaceId != "" {
		existing, err := gormdb.Select[placespg.Place](ctx, databasepg.Query{Limit: 1, Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: req.PlaceId}}},
		}})
		if err != nil {
			return place, err
		}
		if len(existing) == 0 {
			return place, ErrPlaceNotFound
		}
		return existing[0], nil
	}

	if req.Source != placespg.SourceManual && req.ExternalRef != "" {
		existing, err := gormdb.Select[placespg.Place](ctx, byExternalRef(req.Source, req.ExternalRef))
		if err != nil {
			return place, err
		}
		if len(existing) > 0 {
			return existing[0], nil
		}
	} else {
		existing, err := gormdb.Select[placespg.Place](ctx, byNameAndAddress(req.Name, req.Address))
		if err != nil {
			return place, err
		}
		if len(existing) > 0 {
			return existing[0], nil
		}
	}

	source := req.Source
	if source == "" {
		source = placespg.SourceManual
	}

	now := time.Now()
	fresh := placespg.Place{
		Id:               uuid.NewString(),
		Source:           source,
		ExternalRef:      req.ExternalRef,
		Name:             req.Name,
		Category:         req.Category,
		Address:          req.Address,
		Latitude:         req.Latitude,
		Longitude:        req.Longitude,
		Phone:            req.Phone,
		Website:          req.Website,
		OpeningHours:     req.OpeningHours,
		AddedByAccountId: accountId,
		CreateTime:       now,
		UpdateTime:       now,
	}
	if err = gormdb.Create(ctx, &fresh); err != nil {
		return place, err
	}
	return fresh, nil
}

// Recommend resolves (or creates) the target Place, then upserts
// accountId's Recommendation on it -- a repeat recommend just updates the
// comment rather than erroring, since "I already recommended this, let me
// tweak my note" is a normal thing to want to do.
func Recommend(ctx context.Context, accountId string, req placespg.RecommendRequest) (res placespg.PlaceWithStats, err error) {
	ctx = logger.Enter(ctx, "places.Recommend", accountId, req)
	defer logger.Exit(ctx, "places.Recommend", &res, &err)

	place, err := resolvePlace(ctx, accountId, req)
	if err != nil {
		return res, err
	}

	now := time.Now()
	existing, err := gormdb.Select[placespg.Recommendation](ctx, byPlaceAndAccount(place.Id, accountId))
	if err != nil {
		return res, err
	}

	if len(existing) > 0 {
		rec := existing[0]
		rec.Comment = req.Comment
		rec.UpdateTime = now
		if err = gormdb.Update(ctx, &rec, byPlaceAndAccount(place.Id, accountId)); err != nil {
			return res, err
		}
	} else {
		mine, err := gormdb.Select[placespg.Recommendation](ctx, byAccountIdRecommendations(accountId))
		if err != nil {
			return res, err
		}
		if int64(len(mine)) >= parentsubscriptionspg.MaxRecommendationsPerParent {
			return res, ErrRecommendationLimitReached
		}

		rec := placespg.Recommendation{
			Id:         uuid.NewString(),
			PlaceId:    place.Id,
			AccountId:  accountId,
			Comment:    req.Comment,
			CreateTime: now,
			UpdateTime: now,
		}
		if err = gormdb.Create(ctx, &rec); err != nil {
			return res, err
		}
	}

	return GetOne(ctx, accountId, place.Id)
}
