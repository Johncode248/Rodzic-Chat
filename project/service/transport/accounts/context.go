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

func Context(w http.ResponseWriter, r *http.Request) {
	var (
		res accountspg.SelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "accountsgt.Context")
	defer logger.Exit(ctx, "accountsgt.Context", &res, &err)

	userId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || userId == "" {
		http.Error(w, "Error getting userId from context", http.StatusBadRequest)
		return
	}

	res, err = accounts.Select(ctx, accountspg.SelectArgument{
		Query: databasepg.Query{
			Limit: 1,
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
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
