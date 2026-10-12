package caregivers

import (
	"context"
	"strings"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	caregiverspg "github.com/Johncode248/Rodzic-Chat/project/common/caregivers"
)

// SearchFilter narrows the directory. Every field is optional.
type SearchFilter struct {
	Type    caregiverspg.CaregiverType
	Query   string // free text: name, description, subjects, city
	City    string // matches the home city or any served city
	Mode    string // "online" | "stationary" | "travel" | "" (any)
	MaxRate int64  // grosze per hour, 0 = no cap
	Limit   int
	Offset  int
}

// approvedStatuses: 0 is a row from before verification existed.
const approvedSQL = "(review_status = 2 or review_status = 0)"

// Search browses the APPROVED directory: type + optional text/city/mode/rate filters, best
// recommended first, then newest. Pending and rejected profiles are never listed here.
func Search(ctx context.Context, f SearchFilter) (res caregiverspg.ListResponse, err error) {
	where := []string{"c.type = ?", approvedSQL}
	args := []interface{}{int(f.Type)}
	if q := strings.TrimSpace(f.Query); q != "" {
		like := "%" + q + "%"
		where = append(where, "(c.display_name ilike ? or c.description ilike ? or c.subjects ilike ? or c.city ilike ? or c.service_cities ilike ?)")
		args = append(args, like, like, like, like, like)
	}
	if city := strings.TrimSpace(f.City); city != "" {
		like := "%" + city + "%"
		// Online-only providers serve everywhere, so a city search still finds them when asked for online.
		where = append(where, "(c.city ilike ? or c.service_cities ilike ?)")
		args = append(args, like, like)
	}
	switch f.Mode {
	case "online":
		where = append(where, "c.online = true")
	case "stationary":
		where = append(where, "c.stationary = true")
	case "travel":
		where = append(where, "c.travel = true")
	}
	if f.MaxRate > 0 {
		where = append(where, "(c.hourly_rate = 0 or c.hourly_rate <= ?)")
		args = append(args, f.MaxRate)
	}
	sql := `select c.* from caregivers c
 left join (select caregiver_id, count(*) n from caregiver_recommendations group by caregiver_id) r on r.caregiver_id = c.id
 where ` + strings.Join(where, " and ") + ` order by coalesce(r.n, 0) desc, c.create_time desc`
	if f.Limit > 0 {
		sql += " limit ? offset ?"
		args = append(args, f.Limit, f.Offset)
	}
	rows, err := gormdb.CustomSelect[caregiverspg.Caregiver](ctx, sql, args...)
	if err != nil {
		return res, err
	}
	summaries := make([]caregiverspg.Summary, 0, len(rows))
	for _, row := range rows {
		count, countErr := recommendationCount(ctx, row.Id)
		if countErr != nil {
			return res, countErr
		}
		summaries = append(summaries, caregiverspg.Summary{Caregiver: row, RecommendationCount: count})
	}
	return caregiverspg.ListResponse{Caregivers: summaries}, nil
}

// ListMine returns every profile the caller owns, whatever its review status.
func ListMine(ctx context.Context, accountId string) (caregiverspg.ListResponse, error) {
	rows, err := gormdb.CustomSelect[caregiverspg.Caregiver](ctx, "select * from caregivers where account_id = ? order by type", accountId)
	if err != nil {
		return caregiverspg.ListResponse{}, err
	}
	out := make([]caregiverspg.Summary, 0, len(rows))
	for _, r := range rows {
		n, _ := recommendationCount(ctx, r.Id)
		out = append(out, caregiverspg.Summary{Caregiver: r, RecommendationCount: n})
	}
	return caregiverspg.ListResponse{Caregivers: out}, nil
}
