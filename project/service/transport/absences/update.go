package absencesgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	absencespg "github.com/Johncode248/Rodzic-Chat/project/common/absences"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/absences"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Update edits an absence identified by ?id=. Same ownership rule as
// Delete: staff may edit any absence in their organization, a parent only
// one they reported themselves.
func Update(w http.ResponseWriter, r *http.Request) {
	var (
		request absencespg.UpdateRequest
		res     absencespg.UpdateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "absencesgt.Update", r.Body)
	defer logger.Exit(ctx, "absencesgt.Update", &res, &err)

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

	accountType, err := credentials.GetAccountTypeFromContext(r.Context())
	if err != nil {
		http.Error(w, "Error getting accountType from context", http.StatusForbidden)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing 'id' query param", http.StatusBadRequest)
		return
	}

	query := databasepg.Query{
		Groups: []databasepg.Group{
			{
				Filters: []databasepg.Filter{
					{Key: "id", Operator: "=", Value: id},
				},
			},
		},
	}
	query = query.AddOrganizationAuth(organizationId)

	isStaff := accountType == accountspg.AccountTypeDirector.Ordinal() || accountType == accountspg.AccountTypeCoDirector.Ordinal() || accountType == accountspg.AccountTypeTeacher.Ordinal()
	if !isStaff {
		query.Groups[0].Filters = append(query.Groups[0].Filters, databasepg.Filter{
			Key: "reported_by", Operator: "=", Value: accountId,
		})
	}

	res, err = absences.Update(ctx, absencespg.UpdateArgument{
		Query:    query,
		DateFrom: request.DateFrom,
		DateTo:   request.DateTo,
		Reason:   request.Reason,
	})
	if err != nil {
		status := http.StatusInternalServerError
		if err == absences.ErrInvalidDateRange {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}
	if res.Absence.Id == "" {
		http.Error(w, errNotAbsenceOwner.Error(), http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(res)
}
