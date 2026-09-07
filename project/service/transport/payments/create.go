package paymentsgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	paymentspg "github.com/Johncode248/Rodzic-Chat/project/common/payments"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/payments"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Create handles POST /private/payments (director/co-director only, gated
// in module.go): one request, N recipients, N independently-payable
// Payment rows sharing a batch id -- see Payment's doc comment.
func Create(w http.ResponseWriter, r *http.Request) {
	var (
		request paymentspg.CreateRequest
		res     paymentspg.CreateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "paymentsgt.Create", r.Body)
	defer logger.Exit(ctx, "paymentsgt.Create", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	createdBy, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || createdBy == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	for _, recipient := range request.Recipients {
		childOk, err := childInOrganization(ctx, recipient.ChildId, organizationId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !childOk {
			http.Error(w, errChildNotInOrganization.Error(), http.StatusForbidden)
			return
		}

		parentOk, err := parentInOrganization(ctx, recipient.ParentId, organizationId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !parentOk {
			http.Error(w, errParentNotInOrganization.Error(), http.StatusForbidden)
			return
		}
	}

	res, err = payments.Create(ctx, organizationId, createdBy, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
