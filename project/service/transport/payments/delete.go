package paymentsgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	paymentspg "github.com/Johncode248/Rodzic-Chat/project/common/payments"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/payments"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Delete handles DELETE /private/payments?id= (director/co-director only,
// gated in module.go) -- only while the payment is still Unpaid, see
// payments.Delete's doc comment.
func Delete(w http.ResponseWriter, r *http.Request) {
	var (
		res paymentspg.DeleteResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "paymentsgt.Delete")
	defer logger.Exit(ctx, "paymentsgt.Delete", &res, &err)

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

	res, err = payments.Delete(ctx, id, organizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(res)
}
