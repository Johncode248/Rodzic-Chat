package coparents

import (
	"context"
	"errors"
	"sort"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	childeventspg "github.com/Johncode248/Rodzic-Chat/project/common/childevents"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	coparentspg "github.com/Johncode248/Rodzic-Chat/project/common/coparents"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
)

var (
	ErrNotACoParent = errors.New("that parent is not linked to any of your children")
	ErrSelf         = errors.New("you cannot disconnect yourself")
)

func q(key, op string, value interface{}) databasepg.Query {
	return databasepg.Query{Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: key, Operator: op, Value: value}}}}}
}

func links(ctx context.Context, key, op string, value interface{}) ([]parentchildpg.ParentChild, error) {
	return gormdb.Select[parentchildpg.ParentChild](ctx, q(key, op, value))
}

// List returns every other parent sharing at least one child with accountId, with the shared children.
func List(ctx context.Context, accountId string) (res coparentspg.ListResponse, err error) {
	ctx = logger.Enter(ctx, "coparents.List", accountId)
	defer logger.Exit(ctx, "coparents.List", &res, &err)

	res.CoParents = []coparentspg.CoParent{}
	mine, err := links(ctx, "parent_id", "=", accountId)
	if err != nil || len(mine) == 0 {
		return res, err
	}
	childIds := make([]string, 0, len(mine))
	for _, l := range mine {
		childIds = append(childIds, l.ChildId)
	}
	all, err := links(ctx, "child_id", "IN", childIds)
	if err != nil {
		return res, err
	}
	shared := map[string][]string{}
	for _, l := range all {
		if l.ParentId != accountId {
			shared[l.ParentId] = append(shared[l.ParentId], l.ChildId)
		}
	}
	if len(shared) == 0 {
		return res, nil
	}

	parentIds := make([]string, 0, len(shared))
	for id := range shared {
		parentIds = append(parentIds, id)
	}
	accounts, err := gormdb.Select[accountspg.Account](ctx, q("id", "IN", parentIds))
	if err != nil {
		return res, err
	}
	kids, err := gormdb.Select[childrenpg.Child](ctx, q("id", "IN", childIds))
	if err != nil {
		return res, err
	}
	kidName := map[string]string{}
	for _, k := range kids {
		kidName[k.Id] = k.FirstName
	}
	for _, a := range accounts {
		cp := coparentspg.CoParent{Id: a.Id, Name: a.Name, Email: a.Email, SharedChildIds: shared[a.Id]}
		for _, id := range shared[a.Id] {
			cp.SharedChildNames = append(cp.SharedChildNames, kidName[id])
		}
		res.CoParents = append(res.CoParents, cp)
	}
	sort.Slice(res.CoParents, func(i, j int) bool { return res.CoParents[i].Name < res.CoParents[j].Name })
	return res, nil
}

// Disconnect removes coParentId's parent links to every child accountId ALSO parents, and drops
// their shared attendance in each other's personal calendar events, so neither side keeps seeing
// the other's events. The co-parent's other (non-shared) children and their own account are untouched.
func Disconnect(ctx context.Context, accountId, coParentId string) (res coparentspg.DisconnectResponse, err error) {
	ctx = logger.Enter(ctx, "coparents.Disconnect", accountId, coParentId)
	defer logger.Exit(ctx, "coparents.Disconnect", &res, &err)

	res.UnlinkedChildIds = []string{}
	if coParentId == accountId {
		return res, ErrSelf
	}
	mine, err := links(ctx, "parent_id", "=", accountId)
	if err != nil {
		return res, err
	}
	theirs, err := links(ctx, "parent_id", "=", coParentId)
	if err != nil {
		return res, err
	}
	mineSet := map[string]bool{}
	for _, l := range mine {
		mineSet[l.ChildId] = true
	}
	var toRemove []parentchildpg.ParentChild
	for _, l := range theirs {
		if mineSet[l.ChildId] {
			toRemove = append(toRemove, l)
		}
	}
	if len(toRemove) == 0 {
		return res, ErrNotACoParent
	}

	for _, l := range toRemove {
		if _, err = gormdb.Delete[parentchildpg.ParentChild](ctx, q("id", "=", l.Id)); err != nil {
			return res, err
		}
		res.UnlinkedChildIds = append(res.UnlinkedChildIds, l.ChildId)
	}

	// Calendar: events created by one and attended by the other lose the other as an attendee.
	for _, pair := range [][2]string{{accountId, coParentId}, {coParentId, accountId}} {
		creator, other := pair[0], pair[1]
		created, selErr := gormdb.Select[childeventspg.ChildEvent](ctx, q("created_by_account_id", "=", creator))
		if selErr != nil {
			return res, selErr
		}
		for _, e := range created {
			if err = gormdb.GormDB.WithContext(ctx).Exec(
				"DELETE FROM child_event_parents WHERE event_id = ? AND account_id = ?", e.Id, other).Error; err != nil {
				return res, err
			}
		}
	}
	return res, nil
}
