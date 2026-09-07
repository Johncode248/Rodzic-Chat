package paymentsgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	paymentspg "github.com/Johncode248/Rodzic-Chat/project/common/payments"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/payments"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Update handles PATCH /private/payments?id= (director/co-director only,
// gated in module.go) -- edit an unpaid payment's details, or change its
// status (e.g. mark a bank-transfer payment Reconciled). See
// payments.Update's doc comment for the exact rules.
func Update(w http.ResponseWriter, r *http.Request) {
	var (
		request paymentspg.UpdateRequest
		res     paymentspg.UpdateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "paymentsgt.Update", r.Body)
	defer logger.Exit(ctx, "paymentsgt.Update", &res, &err)

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing 'id' query param", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

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

	res, err = payments.Update(ctx, id, organizationId, accountId, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(res)
}
