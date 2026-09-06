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

func Delete(w http.ResponseWriter, r *http.Request) {
	var (
		res accountspg.DeleteResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "accountsgt.Delete")
	defer logger.Exit(ctx, "accountsgt.Delete", &res, &err)

	userId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil {
		http.Error(w, "Auth went wrong", http.StatusForbidden)
		return
	}

	res, err = accounts.Delete(ctx, accountspg.DeleteArgument{
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
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
