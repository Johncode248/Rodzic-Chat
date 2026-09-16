package credentials

import (
	"encoding/json"
	"errors"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	credentialspg "github.com/Johncode248/Rodzic-Chat/project/common/credentials"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/credentials"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/sessions"
	"golang.org/x/crypto/bcrypt"
)

var errCurrentPasswordRequired = errors.New("current password is required to change password")
var errCurrentPasswordIncorrect = errors.New("current password is incorrect")

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

	// Changing the password requires proving you know the CURRENT one --
	// without this, a stolen/leaked access token (XSS, a shared device, a
	// token grabbed off the wire) let an attacker permanently take over the
	// account by setting a new password, with no way for the real owner to
	// prove they're the one asking. Skipped only when the account has no
	// Credential row yet (OAuth-only account setting a password for the
	// first time) -- there's nothing to verify against in that case.
	if request.Password != nil {
		existing, selectErr := credentials.Select(ctx, credentialspg.SelectArgument{
			Query: databasepg.Query{
				Limit: 1,
				Groups: []databasepg.Group{
					{Filters: []databasepg.Filter{{Key: "user_id", Operator: "=", Value: userId}}},
				},
			},
		})
		if selectErr != nil {
			http.Error(w, selectErr.Error(), http.StatusInternalServerError)
			return
		}
		if len(existing.Credentials) > 0 {
			if request.CurrentPassword == nil {
				http.Error(w, errCurrentPasswordRequired.Error(), http.StatusForbidden)
				return
			}
			if bcrypt.CompareHashAndPassword([]byte(existing.Credentials[0].Password), []byte(*request.CurrentPassword)) != nil {
				http.Error(w, errCurrentPasswordIncorrect.Error(), http.StatusForbidden)
				return
			}
		}
	}

	res, err = credentials.Update(ctx, credentialspg.UpdateArgument{
		UserId: userId,
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

	// A password change is exactly the security event that should force
	// every device to re-authenticate -- same reasoning as
	// ConsumePasswordReset. Best-effort: don't fail the response over it,
	// but do log it (RevokeAllForAccount already logs internally via
	// logger.Enter/Exit).
	if request.Password != nil {
		_ = sessions.RevokeAllForAccount(ctx, userId)
	}

	json.NewEncoder(w).Encode(res)
}
