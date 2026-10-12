package caregivers

import (
	"context"
	"errors"
	"strings"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	caregiverspg "github.com/Johncode248/Rodzic-Chat/project/common/caregivers"
)

var ErrNotFound = errors.New("caregiver not found")

func getMine(ctx context.Context, accountId string, caregiverType caregiverspg.CaregiverType) (*caregiverspg.Caregiver, error) {
	rows, err := gormdb.Select[caregiverspg.Caregiver](ctx, databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{
				{Key: "account_id", Operator: "=", Value: accountId},
				{Key: "type", Operator: "=", Value: int(caregiverType)},
			}},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

// GetMine returns the caller's own profile for a type, or ErrNotFound if
// they haven't created one -- used to prefill the edit form / decide
// whether the UI shows "Dodaj swój profil" or "Edytuj swój profil".
func GetMine(ctx context.Context, accountId string, caregiverType caregiverspg.CaregiverType) (res caregiverspg.Caregiver, err error) {
	existing, err := getMine(ctx, accountId, caregiverType)
	if err != nil {
		return res, err
	}
	if existing == nil {
		return res, ErrNotFound
	}
	return *existing, nil
}

func getById(ctx context.Context, id string) (*caregiverspg.Caregiver, error) {
	rows, err := gormdb.Select[caregiverspg.Caregiver](ctx, databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: id}}},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

func recommendationCount(ctx context.Context, caregiverId string) (int64, error) {
	rows, err := gormdb.Select[caregiverspg.CaregiverRecommendation](ctx, databasepg.Query{
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "caregiver_id", Operator: "=", Value: caregiverId}}},
		},
	})
	if err != nil {
		return 0, err
	}
	return int64(len(rows)), nil
}

// GetDetail fetches the aggregate view for one caregiver: the row, its
// recommendation tally, and the two viewer-relative flags (see
// caregiverspg.Detail's doc comment).
func GetDetail(ctx context.Context, id string, callerAccountId string) (res caregiverspg.Detail, err error) {
	caregiver, err := getById(ctx, id)
	if err != nil {
		return res, err
	}
	if caregiver == nil {
		return res, ErrNotFound
	}
	// Pending / rejected profiles are visible to their owner only.
	if caregiver.AccountId != callerAccountId && caregiver.ReviewStatus != caregiverspg.ReviewStatusApproved && caregiver.ReviewStatus != caregiverspg.ReviewStatusUnknown {
		return res, ErrNotFound
	}

	count, err := recommendationCount(ctx, id)
	if err != nil {
		return res, err
	}

	recommended := false
	if callerAccountId != "" {
		recs, recErr := gormdb.Select[caregiverspg.CaregiverRecommendation](ctx, databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{
					{Key: "caregiver_id", Operator: "=", Value: id},
					{Key: "account_id", Operator: "=", Value: callerAccountId},
				}},
			},
		})
		if recErr != nil {
			return res, recErr
		}
		recommended = len(recs) > 0
	}

	recommenders, _ := listRecommenders(ctx, id)

	return caregiverspg.Detail{
		Recommenders:        recommenders,
		Caregiver:           *caregiver,
		RecommendationCount: count,
		RecommendedByMe:     recommended,
		IsMine:              caregiver.AccountId == callerAccountId,
	}, nil
}

// listRecommenders returns the latest 30 endorsements with the endorser's name.
func listRecommenders(ctx context.Context, caregiverId string) ([]caregiverspg.Recommender, error) {
	type row struct {
		Name       string
		Comment    string
		CreateTime time.Time
	}
	rows, err := gormdb.CustomSelect[row](ctx, `select a.name as name, r.comment as comment, r.create_time as create_time
 from caregiver_recommendations r join accounts a on a.id = r.account_id
 where r.caregiver_id = ? order by r.create_time desc limit 30`, caregiverId)
	if err != nil {
		return nil, err
	}
	out := make([]caregiverspg.Recommender, 0, len(rows))
	for _, r := range rows {
		out = append(out, caregiverspg.Recommender{Name: firstName(r.Name), Comment: r.Comment, CreateTime: r.CreateTime})
	}
	return out, nil
}

// firstName shows "Anna K." style names: first name plus the initial of the rest.
func firstName(full string) string {
	parts := strings.Fields(full)
	switch len(parts) {
	case 0:
		return "Rodzic"
	case 1:
		return parts[0]
	}
	return parts[0] + " " + string([]rune(parts[len(parts)-1])[:1]) + "."
}
