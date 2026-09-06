package teacherinvitesgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	teacherinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/teacherinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/teacherinvites"
	"github.com/gorilla/mux"
)

// PublicInfo is unauthenticated — anyone holding the invite's token can see
// which placówka/group it's for before providing any personal information.
func PublicInfo(w http.ResponseWriter, r *http.Request) {
	var (
		res teacherinvitespg.PublicInfoResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "teacherinvitesgt.PublicInfo")
	defer logger.Exit(ctx, "teacherinvitesgt.PublicInfo", &res, &err)

	token := mux.Vars(r)["token"]
	if token == "" {
		http.Error(w, "Missing token", http.StatusBadRequest)
		return
	}

	res, err = teacherinvites.GetPublicInfo(ctx, token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
