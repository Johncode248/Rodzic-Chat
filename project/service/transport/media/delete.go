package mediagt

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/api_bucket"
	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	mediapg "github.com/Johncode248/Rodzic-Chat/project/common/media"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/media"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

var errNotMediaOwner = errors.New("only the uploader or staff may delete this media")

// Delete removes a gallery item identified by ?id=, deleting both the DB
// row and the underlying R2 object. The uploader may delete their own
// upload; staff (director/teacher) may delete anyone's.
func Delete(w http.ResponseWriter, r *http.Request) {
	var (
		res mediapg.DeleteResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "mediagt.Delete")
	defer logger.Exit(ctx, "mediagt.Delete", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(ctx)
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	accountType, err := credentials.GetAccountTypeFromContext(ctx)
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
			Key: "uploader_id", Operator: "=", Value: accountId,
		})
	}

	res, err = media.Delete(ctx, mediapg.DeleteArgument{Query: query})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(res.Media) == 0 {
		http.Error(w, errNotMediaOwner.Error(), http.StatusForbidden)
		return
	}

	for _, item := range res.Media {
		_ = api_bucket.DeleteFile(ctx, bucketName, item.ObjectKey)
	}

	json.NewEncoder(w).Encode(res)
}
