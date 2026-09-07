package paymentsgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	paymentspg "github.com/Johncode248/Rodzic-Chat/project/common/payments"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/payments"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Checkout handles POST /private/payments/checkout?id= -- any authenticated
// caller may call it, but internal/payments.CreateCheckoutSession rejects
// anyone other than the payment's own assigned parent (ErrNotYourPayment),
// so this isn't staff-gated in module.go the way create/update/delete are.
func Checkout(w http.ResponseWriter, r *http.Request) {
	var (
		res paymentspg.CheckoutResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "paymentsgt.Checkout")
	defer logger.Exit(ctx, "paymentsgt.Checkout", &res, &err)

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing 'id' query param", http.StatusBadRequest)
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

	res, err = payments.CreateCheckoutSession(ctx, id, organizationId, accountId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(res)
}
