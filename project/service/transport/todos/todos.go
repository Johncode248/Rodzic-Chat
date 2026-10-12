package todosgt

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	todospg "github.com/Johncode248/Rodzic-Chat/project/common/todos"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/todos"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

func account(w http.ResponseWriter, r *http.Request) (string, bool) {
	id, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || id == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return "", false
	}
	return id, true
}

func fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, todos.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, todos.ErrTitleRequired), errors.Is(err, todos.ErrTitleTooLong), errors.Is(err, todos.ErrNoteTooLong), errors.Is(err, todos.ErrRepeat), errors.Is(err, todos.ErrChild), errors.Is(err, todos.ErrKey):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// List handles GET /private/todos.
func List(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Enter(r.Context(), "todosgt.List")
	var err error
	defer logger.Exit(ctx, "todosgt.List", &err)
	accountId, ok := account(w, r)
	if !ok {
		return
	}
	res, err := todos.List(ctx, accountId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(res)
}

// Create handles POST /private/todos.
func Create(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Enter(r.Context(), "todosgt.Create")
	var err error
	defer logger.Exit(ctx, "todosgt.Create", &err)
	accountId, ok := account(w, r)
	if !ok {
		return
	}
	var req todospg.CreateRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}
	res, err := todos.Create(ctx, accountId, req)
	if err != nil {
		fail(w, err)
		return
	}
	json.NewEncoder(w).Encode(res)
}

// Update handles PATCH /private/todos?id=.
func Update(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Enter(r.Context(), "todosgt.Update")
	var err error
	defer logger.Exit(ctx, "todosgt.Update", &err)
	accountId, ok := account(w, r)
	if !ok {
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}
	var req todospg.UpdateRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}
	res, err := todos.Update(ctx, accountId, id, req)
	if err != nil {
		fail(w, err)
		return
	}
	json.NewEncoder(w).Encode(res)
}

// Delete handles DELETE /private/todos?id=.
func Delete(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Enter(r.Context(), "todosgt.Delete")
	var err error
	defer logger.Exit(ctx, "todosgt.Delete", &err)
	accountId, ok := account(w, r)
	if !ok {
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}
	if err = todos.Delete(ctx, accountId, id); err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Snooze handles POST /private/todos/snooze {"key":"payment:...","until":RFC3339} -- hides an automatic to-do.
func Snooze(w http.ResponseWriter, r *http.Request) {
	accountId, ok := account(w, r)
	if !ok {
		return
	}
	var req todospg.SnoozeRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}
	if err := todos.Snooze(r.Context(), accountId, req); err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Unsnooze handles DELETE /private/todos/snooze?key= .
func Unsnooze(w http.ResponseWriter, r *http.Request) {
	accountId, ok := account(w, r)
	if !ok {
		return
	}
	if err := todos.Unsnooze(r.Context(), accountId, r.URL.Query().Get("key")); err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetPrefs handles GET /private/todos/prefs; SetPrefs handles PUT /private/todos/prefs {"digest":bool,"due_reminders":bool}.
func GetPrefs(w http.ResponseWriter, r *http.Request) {
	accountId, ok := account(w, r)
	if !ok {
		return
	}
	json.NewEncoder(w).Encode(todos.GetPrefs(r.Context(), accountId))
}

func SetPrefs(w http.ResponseWriter, r *http.Request) {
	accountId, ok := account(w, r)
	if !ok {
		return
	}
	var req struct {
		Digest       bool `json:"digest"`
		DueReminders bool `json:"due_reminders"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}
	res, err := todos.SetPrefs(r.Context(), accountId, req.Digest, req.DueReminders)
	if err != nil {
		fail(w, err)
		return
	}
	json.NewEncoder(w).Encode(res)
}
