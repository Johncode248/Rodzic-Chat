package registrationlinksgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	registrationlinkspg "github.com/Johncode248/Rodzic-Chat/project/common/registrationlinks"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/registrationlinks"
	"github.com/gorilla/mux"
)

// PublicInfo is unauthenticated — anyone holding the link's token can see
// which placówka/group it's for and pick their child, before providing any
// personal information.
func PublicInfo(w http.ResponseWriter, r *http.Request) {
	var (
		res registrationlinkspg.PublicInfoResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "registrationlinksgt.PublicInfo")
	defer logger.Exit(ctx, "registrationlinksgt.PublicInfo", &res, &err)

	token := mux.Vars(r)["token"]
	if token == "" {
		http.Error(w, "Missing token", http.StatusBadRequest)
		return
	}

	res, err = registrationlinks.GetPublicInfo(ctx, token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
