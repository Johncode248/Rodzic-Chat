package mediagt

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/Johncode248/Rodzic-Chat/api_bucket"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	mediapg "github.com/Johncode248/Rodzic-Chat/project/common/media"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/media"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
	"github.com/google/uuid"
)

var errOnlyStaffCanAttachToEvent = errors.New("only staff can attach a file to an event")

// bucketName is the shared Cloudflare R2 bucket already provisioned for
// this project (see api_bucket.module.go / the old demo code). Object keys
// are namespaced under rodzic-chat/media/<org>/ so this app's files don't
// collide with anything else already living in that bucket.
const bucketName = "dom2sky"

const maxUploadSize = 25 * 1024 * 1024 // 25MB per photo

// Upload handles POST /private/media as a multipart/form-data upload: a
// "file" field plus optional "event_id" / "group_id" / "caption" form
// fields. The file is pushed to R2 first, then a Media row is created --
// if the DB write fails the just-uploaded object is cleaned up.
func Upload(w http.ResponseWriter, r *http.Request) {
	var (
		res mediapg.CreateResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "mediagt.Upload")
	defer logger.Exit(ctx, "mediagt.Upload", &res, &err)

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

	if err = r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "Error parsing multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// General gallery photos (no event_id, no category) stay open to any
	// org member; only staff may attach a file to a specific Event, or
	// upload a categorized document (e.g. "schedule" / "meal_plan") --
	// per product decision, this is how consent-document and
	// plan-zajęć/plan-obiadowy uploads are restricted without a parallel
	// "document" concept.
	if r.FormValue("event_id") != "" || r.FormValue("category") != "" {
		accountType, err := credentials.GetAccountTypeFromContext(ctx)
		isStaff := err == nil && (accountType == accountspg.AccountTypeDirector.Ordinal() ||
			accountType == accountspg.AccountTypeCoDirector.Ordinal() ||
			accountType == accountspg.AccountTypeTeacher.Ordinal())
		if !isStaff {
			http.Error(w, errOnlyStaffCanAttachToEvent.Error(), http.StatusForbidden)
			return
		}
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Missing 'file' field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	objectKey := fmt.Sprintf("rodzic-chat/media/%s/%s%s", organizationId, uuid.NewString(), filepath.Ext(header.Filename))
	contentType := header.Header.Get("Content-Type")

	if err = api_bucket.UploadFile(ctx, bucketName, objectKey, file, header.Size, contentType); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err = media.Create(ctx, mediapg.CreateArgument{
		OrganizationId:   organizationId,
		EventId:          r.FormValue("event_id"),
		GroupId:          r.FormValue("group_id"),
		UploaderId:       accountId,
		ObjectKey:        objectKey,
		OriginalFilename: header.Filename,
		ContentType:      contentType,
		Caption:          r.FormValue("caption"),
		Category:         r.FormValue("category"),
	})
	if err != nil {
		_ = api_bucket.DeleteFile(ctx, bucketName, objectKey)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
