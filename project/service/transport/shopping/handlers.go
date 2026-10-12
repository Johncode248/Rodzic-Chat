package shoppinggt

import (
	"encoding/json"
	"errors"
	"net/http"

	sp "github.com/Johncode248/Rodzic-Chat/project/common/shopping"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/shopping"
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
	case errors.Is(err, shopping.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, shopping.ErrInvalid):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func reply(w http.ResponseWriter, v any, err error) {
	if err != nil {
		fail(w, err)
		return
	}
	json.NewEncoder(w).Encode(v)
}

func decode[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return v, false
	}
	return v, true
}

func noContent(w http.ResponseWriter, err error) {
	if err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /private/shopping/lists?status=active|completed
func Lists(w http.ResponseWriter, r *http.Request) {
	if a, ok := account(w, r); ok {
		status := r.URL.Query().Get("status")
		if status == "" {
			status = sp.StatusActive
		}
		v, err := shopping.Lists(r.Context(), a, status)
		reply(w, v, err)
	}
}

func CreateList(w http.ResponseWriter, r *http.Request) {
	if a, ok := account(w, r); ok {
		if req, ok := decode[sp.CreateListRequest](w, r); ok {
			v, err := shopping.CreateList(r.Context(), a, req)
			reply(w, v, err)
		}
	}
}

func UpdateList(w http.ResponseWriter, r *http.Request) {
	if a, ok := account(w, r); ok {
		if req, ok := decode[sp.UpdateListRequest](w, r); ok {
			v, err := shopping.UpdateList(r.Context(), a, r.URL.Query().Get("id"), req)
			reply(w, v, err)
		}
	}
}

func DeleteList(w http.ResponseWriter, r *http.Request) {
	if a, ok := account(w, r); ok {
		noContent(w, shopping.DeleteList(r.Context(), a, r.URL.Query().Get("id")))
	}
}

func CopyList(w http.ResponseWriter, r *http.Request) {
	if a, ok := account(w, r); ok {
		var req sp.CopyListRequest
		_ = json.NewDecoder(r.Body).Decode(&req) // optional body
		v, err := shopping.CopyList(r.Context(), a, r.URL.Query().Get("id"), req)
		reply(w, v, err)
	}
}

// GET /private/shopping/list?id=
func Detail(w http.ResponseWriter, r *http.Request) {
	if a, ok := account(w, r); ok {
		v, err := shopping.Detail(r.Context(), a, r.URL.Query().Get("id"))
		reply(w, v, err)
	}
}

func AddItems(w http.ResponseWriter, r *http.Request) {
	if a, ok := account(w, r); ok {
		if req, ok := decode[sp.AddItemsRequest](w, r); ok {
			v, err := shopping.AddItems(r.Context(), a, req)
			reply(w, v, err)
		}
	}
}

func UpdateItem(w http.ResponseWriter, r *http.Request) {
	if a, ok := account(w, r); ok {
		if req, ok := decode[sp.UpdateItemRequest](w, r); ok {
			v, err := shopping.UpdateItem(r.Context(), a, r.URL.Query().Get("id"), req)
			reply(w, v, err)
		}
	}
}

func DeleteItem(w http.ResponseWriter, r *http.Request) {
	if a, ok := account(w, r); ok {
		noContent(w, shopping.DeleteItem(r.Context(), a, r.URL.Query().Get("id")))
	}
}

// GET /private/shopping/history?q=
func History(w http.ResponseWriter, r *http.Request) {
	if a, ok := account(w, r); ok {
		reply(w, shopping.History(r.Context(), a, r.URL.Query().Get("q")), nil)
	}
}
