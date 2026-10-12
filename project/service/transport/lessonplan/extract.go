package lessonplangt

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/aichat"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// maxImageSize mirrors project/service/transport/media/upload.go's per-file cap -- a phone photo
// of a paper plan zajęć comfortably fits well under this.
const maxImageSize = 10 * 1024 * 1024

// Extract handles POST /private/lesson-plan/extract -- a parent (or staff) photographs a paper
// plan zajęć and this sends it to Claude (vision) to read back a structured weekly schedule (see
// project/service/internal/aichat.ExtractScheduleFromImage). Client shows the result for the
// parent to confirm/correct before anything is saved -- this endpoint itself does not create any
// calendar events, it only reads the photo.
//
// No usage/quota gate here yet -- same caveat as project/service/transport/aichat.Send's doc
// comment: this spends a real Anthropic API call per request with nothing capping abuse. Revisit
// once real per-parent limits exist; for now this is behind being logged in at all, same as every
// other private endpoint.
func Extract(w http.ResponseWriter, r *http.Request) {
	var (
		res aichat.ScheduleExtraction
		err error
	)
	ctx := logger.Enter(r.Context(), "lessonplangt.Extract")
	defer logger.Exit(ctx, "lessonplangt.Extract", &res, &err)

	if _, err = credentials.GetUserIdFromContext(ctx); err != nil {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxImageSize)
	if err = r.ParseMultipartForm(maxImageSize); err != nil {
		http.Error(w, "Error parsing multipart form: "+err.Error(), http.StatusRequestEntityTooLarge)
		return
	}

	file, header, ferr := r.FormFile("file")
	if ferr != nil {
		err = ferr
		http.Error(w, "Missing 'file' field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, rerr := io.ReadAll(file)
	if rerr != nil {
		err = rerr
		http.Error(w, "Error reading file", http.StatusBadRequest)
		return
	}

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "image/jpeg"
	}

	res, err = aichat.ExtractScheduleFromImage(ctx, mimeType, base64.StdEncoding.EncodeToString(data))
	if err != nil {
		if err == aichat.ErrNotConfigured {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
