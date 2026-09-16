package accountsgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/pagination"
)

// Roster lists every account in the caller's organization (staff-only --
// gated by staffOnly in module.go) -- the "who is on my team" view the web
// admin platform needs, distinct from Context's "who am I" self-lookup.
// Optional ?account_type= filters to "parent" or "teacher" or "director".
// Optional ?q= free-text searches name/login/email (case-insensitive,
// substring match). Optional ?status=removed shows only soft-deleted
// accounts instead of active ones (e.g. parents in the 3-month grace
// window before the reaper hard-deletes them) -- the default (no
// ?status=) only ever shows active accounts.
func Roster(w http.ResponseWriter, r *http.Request) {
	var (
		res accountspg.SelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "accountsgt.Roster")
	defer logger.Exit(ctx, "accountsgt.Roster", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	var accountTypeFilter *databasepg.Filter
	if accountType := r.URL.Query().Get("account_type"); accountType != "" {
		mapped, err := authorizationpg.MapAccountType(accountType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		accountTypeFilter = &databasepg.Filter{
			Key: "account_type", Operator: "=", Value: mapped.Ordinal(),
		}
	}

	var query databasepg.Query

	if q := r.URL.Query().Get("q"); q != "" {
		like := "%" + q + "%"
		// One OR-group per searched column (name/login/email) -- each gets
		// the account_type filter (if any) ANDed in below via
		// AddOrganizationAuth's existing "append to every group" behavior.
		query.Groups = []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "name", Operator: "ILIKE", Value: like}}},
			{Filters: []databasepg.Filter{{Key: "login", Operator: "ILIKE", Value: like}}},
			{Filters: []databasepg.Filter{{Key: "email", Operator: "ILIKE", Value: like}}},
		}
		if accountTypeFilter != nil {
			for i := range query.Groups {
				query.Groups[i].Filters = append(query.Groups[i].Filters, *accountTypeFilter)
			}
		}
	} else {
		query.Groups = []databasepg.Group{{}}
		if accountTypeFilter != nil {
			query.Groups[0].Filters = append(query.Groups[0].Filters, *accountTypeFilter)
		}
	}

	removed := r.URL.Query().Get("status") == "removed"
	if removed {
		for i := range query.Groups {
			query.Groups[i].Filters = append(query.Groups[i].Filters, databasepg.Filter{
				Key: "deleted_at", Operator: "IS NOT", Value: nil,
			})
		}
	}

	query = query.AddOrganizationAuth(organizationId)
	query.OrderBy = databasepg.Order{Key: "name", Direction: databasepg.DirectionAscending}
	query.Limit, query.Offset = pagination.Parse(r, pagination.DefaultMaxLimit)

	if removed {
		res, err = accounts.SelectUnscoped(ctx, accountspg.SelectArgument{Query: query})
	} else {
		res, err = accounts.Select(ctx, accountspg.SelectArgument{Query: query})
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
