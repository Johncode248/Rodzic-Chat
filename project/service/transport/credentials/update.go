package credentials

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	credentialspg "github.com/Johncode248/Rodzic-Chat/project/common/credentials"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/credentials"
)

func Update(w http.ResponseWriter, r *http.Request) {
	var (
		request credentialspg.UpdateRequest
		res     credentialspg.UpdateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "credentialsgt.Update", r.Body)
	defer logger.Exit(ctx, "credentialsgt.Update", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	userId, err := GetUserIdFromContext(r.Context())
	if err != nil || userId == "" {
		http.Error(w, "Error getting userId from context", http.StatusForbidden)
		return
	}

	res, err = credentials.Update(ctx, credentialspg.UpdateArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{
							Key:      "user_id",
							Operator: "=",
							Value:    userId,
						},
					},
				},
			},
		},
		Password: request.Password,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
