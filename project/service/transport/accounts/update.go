package accountsgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

func Update(w http.ResponseWriter, r *http.Request) {
	var (
		request accountspg.UpdateRequest
		res     accountspg.UpdateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "accountsgt.Update", r.Body)
	defer logger.Exit(ctx, "accountsgt.Update", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	userId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || userId == "" {
		http.Error(w, "Error getting userId from context", http.StatusForbidden)
		return
	}

	res, err = accounts.Update(ctx, accountspg.UpdateArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{
							Key:      "id",
							Operator: "=",
							Value:    userId,
						},
					},
				},
			},
		},
		Name:  request.Name,
		Phone: request.Phone,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
