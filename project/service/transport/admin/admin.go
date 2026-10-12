// Package admingt is the LOCAL platform-admin surface ("centrum dowodzenia"): JSON endpoints under
// /admin/api that see and change everything, with NO login. That is only acceptable because
// Mount refuses to serve anything unless ADMIN_MODE=1 AND the request comes from the machine
// itself (loopback, no proxy headers). It is meant to be run on a developer laptop and is never
// mounted in production: the Mount call in transport/module.go stays commented out.
package admingt

import (
	"encoding/json"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// Mount registers /admin/api/* on the router.
func Mount(router *mux.Router) {
	if os.Getenv("ADMIN_MODE") != "1" {
		return
	}
	api := router.PathPrefix("/admin/api").Subrouter()
	api.Use(guard)

	api.HandleFunc("/ping", ping).Methods("GET")
	api.HandleFunc("/stats", stats).Methods("GET")

	api.HandleFunc("/tables", tables).Methods("GET")
	api.HandleFunc("/t/{table}", listRows).Methods("GET")
	api.HandleFunc("/t/{table}", createRow).Methods("POST")
	api.HandleFunc("/t/{table}", updateRow).Methods("PATCH")
	api.HandleFunc("/t/{table}", deleteRow).Methods("DELETE")
	api.HandleFunc("/t/{table}/get", getRow).Methods("GET")

	api.HandleFunc("/accounts/{id}/overview", accountOverview).Methods("GET")
	api.HandleFunc("/accounts/{id}/status", setAccountStatus).Methods("POST")

	api.HandleFunc("/caregivers", caregiversList).Methods("GET")
	api.HandleFunc("/caregivers/decide", caregiversDecide).Methods("POST")

	api.HandleFunc("/moderation/sources", moderationSources).Methods("GET")
	api.HandleFunc("/moderation", moderationFeed).Methods("GET")
	api.HandleFunc("/moderation/remove", moderationRemove).Methods("POST")

	api.HandleFunc("/push/audience", pushAudience).Methods("POST")
	api.HandleFunc("/push/send", pushSend).Methods("POST")
}

// guard: loopback only. A reverse proxy in front of the process would make every request look
// like loopback, so any forwarding header is treated as "not local" too.
func guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		ip := net.ParseIP(host)
		if err != nil || ip == nil || !ip.IsLoopback() ||
			r.Header.Get("X-Forwarded-For") != "" || r.Header.Get("X-Real-Ip") != "" || r.Header.Get("Forwarded") != "" {
			http.Error(w, `{"error":"admin API is local-only"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func ping(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "mode": "admin"})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]any{"error": msg})
}

func readBody(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		fail(w, http.StatusBadRequest, "niepoprawny JSON: "+err.Error())
		return false
	}
	return true
}
