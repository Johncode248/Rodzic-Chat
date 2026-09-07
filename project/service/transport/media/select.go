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
// filtered by ?event_id= and/or ?group_id=. When ?group_id= is given
// without ?event_id= (the general-gallery case -- a parent browsing their
// child's group), org-wide items (no group_id of their own, e.g. a
// placówka-wide "plan obiadowy") are included alongside that group's items,
// since both are meant to be visible to that parent. An ?event_id= query
// stays an exact match regardless of ?group_id=, since a consent file is
// tied to one specific event, not an audience.
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

	eventId := r.URL.Query().Get("event_id")
	groupId := r.URL.Query().Get("group_id")

	var query databasepg.Query

	if groupId != "" && eventId == "" {
		query.Groups = []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "group_id", Operator: "=", Value: groupId}}},
			{Filters: []databasepg.Filter{{Key: "group_id", Operator: "=", Value: ""}, {Key: "event_id", Operator: "=", Value: ""}}},
		}
	} else {
		query.Groups = []databasepg.Group{{}}
		if eventId != "" {
			query.Groups[0].Filters = append(query.Groups[0].Filters, databasepg.Filter{
				Key: "event_id", Operator: "=", Value: eventId,
			})
		}
		if groupId != "" {
			query.Groups[0].Filters = append(query.Groups[0].Filters, databasepg.Filter{
				Key: "group_id", Operator: "=", Value: groupId,
			})
		}
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
