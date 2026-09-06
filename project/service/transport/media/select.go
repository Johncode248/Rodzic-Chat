package mediagt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/api_bucket"
	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	mediapg "github.com/Johncode248/Rodzic-Chat/project/common/media"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/media"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// SelectResponse resolves each row's downloadable URL via a 1h presigned
// R2 link -- good enough for a gallery view; a production version might
// want longer-lived or CDN-backed URLs instead.
type SelectResponse struct {
	Media []mediapg.MediaWithUrl `json:"media,omitempty"`
}

// Select lists gallery media in the caller's organization, optionally
// filtered by ?event_id= or ?group_id=.
func Select(w http.ResponseWriter, r *http.Request) {
	var (
		res SelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "mediagt.Select")
	defer logger.Exit(ctx, "mediagt.Select", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(ctx)
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	query := databasepg.Query{
		Groups: []databasepg.Group{{}},
	}

	if eventId := r.URL.Query().Get("event_id"); eventId != "" {
		query.Groups[0].Filters = append(query.Groups[0].Filters, databasepg.Filter{
			Key: "event_id", Operator: "=", Value: eventId,
		})
	}

	if groupId := r.URL.Query().Get("group_id"); groupId != "" {
		query.Groups[0].Filters = append(query.Groups[0].Filters, databasepg.Filter{
			Key: "group_id", Operator: "=", Value: groupId,
		})
	}

	query = query.AddOrganizationAuth(organizationId)

	result, err := media.Select(ctx, mediapg.SelectArgument{Query: query})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, item := range result.Media {
		url, err := api_bucket.DownloadFile(ctx, bucketName, item.ObjectKey, item.OriginalFilename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		res.Media = append(res.Media, mediapg.MediaWithUrl{Media: item, Url: url})
	}

	json.NewEncoder(w).Encode(res)
}
