package helprequestsgt

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	hp "github.com/Johncode248/Rodzic-Chat/project/common/helprequests"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/helprequests"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

func viewer(w http.ResponseWriter, r *http.Request) (string, bool) {
	id, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || id == "" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return "", false
	}
	return id, true
}

func fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, helprequests.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, helprequests.ErrForbidden):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, helprequests.ErrTitle), errors.Is(err, helprequests.ErrCategory), errors.Is(err, helprequests.ErrText),
		errors.Is(err, helprequests.ErrLink), errors.Is(err, helprequests.ErrClosed), errors.Is(err, helprequests.ErrTooManyReplies), errors.Is(err, helprequests.ErrOwnReply),
		errors.Is(err, helprequests.ErrRating), errors.Is(err, helprequests.ErrPhoto), errors.Is(err, helprequests.ErrPhotoSize):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Categories(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(hp.CategoriesResponse{Categories: hp.Categories})
}

// List handles GET /private/requests?q=&category=&city=&scope=&sort=&limit=&offset= (legacy mine=1 == scope=mine).
func List(w http.ResponseWriter, r *http.Request) {
	id, ok := viewer(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	scope := q.Get("scope")
	if q.Get("mine") == "1" {
		scope = "mine"
	}
	res, err := helprequests.List(r.Context(), id, helprequests.ListFilter{Q: q.Get("q"), Category: q.Get("category"), City: q.Get("city"),
		Scope: scope, Sort: q.Get("sort"), Limit: limit, Offset: offset})
	if err != nil {
		fail(w, err)
		return
	}
	json.NewEncoder(w).Encode(res)
}

func Create(w http.ResponseWriter, r *http.Request) {
	id, ok := viewer(w, r)
	if !ok {
		return
	}
	var req hp.CreateRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}
	res, err := helprequests.Create(r.Context(), id, req)
	if err != nil {
		fail(w, err)
		return
	}
	json.NewEncoder(w).Encode(res)
}

func Detail(w http.ResponseWriter, r *http.Request) {
	id, ok := viewer(w, r)
	if !ok {
		return
	}
	res, err := helprequests.Detail(r.Context(), id, r.URL.Query().Get("id"))
	if err != nil {
		fail(w, err)
		return
	}
	json.NewEncoder(w).Encode(res)
}

// Close handles POST /private/requests/close?id=&closed=1|0.
func Close(w http.ResponseWriter, r *http.Request) {
	id, ok := viewer(w, r)
	if !ok {
		return
	}
	if err := helprequests.SetClosed(r.Context(), id, r.URL.Query().Get("id"), r.URL.Query().Get("closed") != "0"); err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := viewer(w, r)
	if !ok {
		return
	}
	if err := helprequests.Delete(r.Context(), id, r.URL.Query().Get("id")); err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func Reply(w http.ResponseWriter, r *http.Request) {
	id, ok := viewer(w, r)
	if !ok {
		return
	}
	var req hp.ReplyRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}
	res, err := helprequests.Reply(r.Context(), id, req)
	if err != nil {
		fail(w, err)
		return
	}
	json.NewEncoder(w).Encode(res)
}

func DeleteReply(w http.ResponseWriter, r *http.Request) {
	id, ok := viewer(w, r)
	if !ok {
		return
	}
	if err := helprequests.DeleteReply(r.Context(), id, r.URL.Query().Get("id")); err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Vote handles POST /private/requests/replies/vote?id= -- toggles "Też polecam" on a reply.
func Vote(w http.ResponseWriter, r *http.Request) {
	id, ok := viewer(w, r)
	if !ok {
		return
	}
	on, err := helprequests.ToggleVote(r.Context(), id, r.URL.Query().Get("id"))
	if err != nil {
		fail(w, err)
		return
	}
	json.NewEncoder(w).Encode(map[string]bool{"voted": on})
}

// Similar handles GET /private/requests/similar?q= -- existing questions resembling a draft title.
func Similar(w http.ResponseWriter, r *http.Request) {
	id, ok := viewer(w, r)
	if !ok {
		return
	}
	res, err := helprequests.Similar(r.Context(), id, r.URL.Query().Get("q"))
	if err != nil {
		fail(w, err)
		return
	}
	json.NewEncoder(w).Encode(res)
}

// Best handles POST /private/requests/best?id=&reply_id= (empty reply_id clears it).
func Best(w http.ResponseWriter, r *http.Request) {
	id, ok := viewer(w, r)
	if !ok {
		return
	}
	if err := helprequests.SetBest(r.Context(), id, r.URL.Query().Get("id"), r.URL.Query().Get("reply_id")); err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PhotoPut handles POST /private/requests/photos: the body is the image; returns {"id": ...}.
func PhotoPut(w http.ResponseWriter, r *http.Request) {
	id, ok := viewer(w, r)
	if !ok {
		return
	}
	photoId, err := helprequests.SavePhoto(r.Context(), id, r.Body)
	if err != nil {
		fail(w, err)
		return
	}
	json.NewEncoder(w).Encode(hp.PhotoResponse{Id: photoId})
}

// PhotoGet handles GET /private/requests/photos?id= -- the raw image (any signed-in account).
func PhotoGet(w http.ResponseWriter, r *http.Request) {
	if _, ok := viewer(w, r); !ok {
		return
	}
	p, err := helprequests.GetPhoto(r.Context(), r.URL.Query().Get("id"))
	if err != nil {
		fail(w, err)
		return
	}
	w.Header().Set("Content-Type", p.ContentType)
	w.Header().Set("Cache-Control", "private, max-age=86400")
	w.Write(p.Data)
}

// Follows handles GET /private/requests/follows; FollowSet handles POST ...?category=&on=1|0.
func Follows(w http.ResponseWriter, r *http.Request) {
	id, ok := viewer(w, r)
	if !ok {
		return
	}
	res, err := helprequests.Follows(r.Context(), id)
	if err != nil {
		fail(w, err)
		return
	}
	json.NewEncoder(w).Encode(res)
}

func FollowSet(w http.ResponseWriter, r *http.Request) {
	id, ok := viewer(w, r)
	if !ok {
		return
	}
	if err := helprequests.SetFollow(r.Context(), id, r.URL.Query().Get("category"), r.URL.Query().Get("on") != "0"); err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Activity handles GET /private/requests/activity -- the viewer's questions with replies they have not seen.
func Activity(w http.ResponseWriter, r *http.Request) {
	id, ok := viewer(w, r)
	if !ok {
		return
	}
	res, err := helprequests.Activity(r.Context(), id)
	if err != nil {
		fail(w, err)
		return
	}
	json.NewEncoder(w).Encode(res)
}

// Stats handles GET /private/requests/stats -- how many parents the viewer has helped.
func Stats(w http.ResponseWriter, r *http.Request) {
	id, ok := viewer(w, r)
	if !ok {
		return
	}
	res, err := helprequests.Stats(r.Context(), id)
	if err != nil {
		fail(w, err)
		return
	}
	json.NewEncoder(w).Encode(res)
}
