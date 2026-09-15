package children

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/crypto"
	"github.com/Johncode248/Rodzic-Chat/logger"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentchild"
)

var ErrNotYourChild = errors.New("you are not a parent of this child")
var ErrChildNotFound = errors.New("child not found")

func byIdQuery(id string) databasepg.Query {
	return databasepg.Query{Limit: 1, Groups: []databasepg.Group{
		{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: id}}},
	}}
}

func getMineOwned(ctx context.Context, accountId string, childId string) (child childrenpg.Child, err error) {
	existing, err := gormdb.Select[childrenpg.Child](ctx, byIdQuery(childId))
	if err != nil {
		return child, err
	}
	if len(existing) == 0 {
		return child, ErrChildNotFound
	}
	child = existing[0]

	isParent, err := parentchild.IsParent(ctx, accountId, childId)
	if err != nil {
		return child, err
	}
	if !isParent {
		return child, ErrNotYourChild
	}

	return child, nil
}

// UpdateMine lets any parent linked to childId edit its own fields --
// name, birth date, notes, PESEL (encrypted before storage, see
// library/crypto). Every linked parent has equal edit rights; there's no
// concept of a "primary" parent here, matching how ParentChild itself has
// no ordering/priority between the (up to two) linked accounts.
func UpdateMine(ctx context.Context, accountId string, childId string, req childrenpg.UpdateMineRequest) (res childrenpg.Child, err error) {
	ctx = logger.Enter(ctx, "children.UpdateMine", accountId, childId, req)
	defer logger.Exit(ctx, "children.UpdateMine", &res, &err)

	child, err := getMineOwned(ctx, accountId, childId)
	if err != nil {
		return res, err
	}

	if req.FirstName != nil {
		child.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		child.LastName = *req.LastName
	}
	if req.BirthDate != nil && *req.BirthDate != "" {
		parsed, parseErr := time.Parse("2006-01-02", *req.BirthDate)
		if parseErr != nil {
			return res, parseErr
		}
		child.BirthDate = parsed
	}
	if req.Notes != nil {
		child.Notes = *req.Notes
	}
	if req.Pesel != nil {
		encrypted, encErr := crypto.EncryptField(*req.Pesel)
		if encErr != nil {
			return res, encErr
		}
		child.EncryptedPesel = encrypted
	}
	child.UpdateTime = time.Now()

	if err = gormdb.Update(ctx, &child, byIdQuery(childId)); err != nil {
		return res, err
	}

	child.HasPesel = child.EncryptedPesel != ""

	return child, nil
}

// RevealPesel decrypts and returns childId's PESEL -- the one place this
// ever leaves ciphertext form, called only when a parent explicitly taps
// the "eye" icon in the UI, never as part of the ordinary children list.
func RevealPesel(ctx context.Context, accountId string, childId string) (res childrenpg.RevealPeselResponse, err error) {
	ctx = logger.Enter(ctx, "children.RevealPesel", accountId, childId)
	defer logger.Exit(ctx, "children.RevealPesel", &res, &err)

	child, err := getMineOwned(ctx, accountId, childId)
	if err != nil {
		return res, err
	}

	plaintext, err := crypto.DecryptField(child.EncryptedPesel)
	if err != nil {
		return res, err
	}

	return childrenpg.RevealPeselResponse{Pesel: plaintext}, nil
}
