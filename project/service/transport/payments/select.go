package paymentsgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	paymentspg "github.com/Johncode248/Rodzic-Chat/project/common/payments"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/payments"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/pagination"
)

// Select lists payments in the caller's organization. Director/co-director
// see every payment; anyone else (parent or teacher) only ever sees
// payments assigned to them as the paying parent -- a teacher has no
// special visibility here, per product direction ("dyrektor i rodzic
// powinni te płatności widzieć"). Optional ?child_id=/?status=/?batch_id=
// narrow further (staff only for ?child_id=, since it's meaningless as a
// self-filter for a parent whose own rows are already scoped).
func Select(w http.ResponseWriter, r *http.Request) {
	var (
		res paymentspg.SelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "paymentsgt.Select")
	defer logger.Exit(ctx, "paymentsgt.Select", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	accountType, err := credentials.GetAccountTypeFromContext(r.Context())
	if err != nil {
		http.Error(w, "Error getting accountType from context", http.StatusForbidden)
		return
	}

	isDirectorLike := accountType == accountspg.AccountTypeDirector.Ordinal() ||
		accountType == accountspg.AccountTypeCoDirector.Ordinal()

	filters := []databasepg.Filter{}

	if !isDirectorLike {
		filters = append(filters, databasepg.Filter{Key: "parent_id", Operator: "=", Value: accountId})
	} else if childId := r.URL.Query().Get("child_id"); childId != "" {
		filters = append(filters, databasepg.Filter{Key: "child_id", Operator: "=", Value: childId})
	}

	if status := r.URL.Query().Get("status"); status != "" {
		mapped := paymentspg.MapPaymentStatus(status)
		if mapped != paymentspg.PaymentStatusUnknown {
			filters = append(filters, databasepg.Filter{Key: "status", Operator: "=", Value: mapped.Ordinal()})
		}
	}

	if batchId := r.URL.Query().Get("batch_id"); batchId != "" {
		filters = append(filters, databasepg.Filter{Key: "batch_id", Operator: "=", Value: batchId})
	}

	query := databasepg.Query{Groups: []databasepg.Group{{Filters: filters}}}
	query = query.AddOrganizationAuth(organizationId)
	query.OrderBy = databasepg.Order{Key: "create_time", Direction: databasepg.DirectionDescending}
	query.Limit, query.Offset = pagination.Parse(r, pagination.DefaultMaxLimit)

	res, err = payments.Select(ctx, query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
