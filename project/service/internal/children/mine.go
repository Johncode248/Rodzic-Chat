package children

import (
	"context"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentchild"
	"github.com/google/uuid"
)

// CreateMine lets a parent add a child entirely on their own -- no
// placowka, no invite link, no director involved. The resulting Child has
// an empty OrganizationId/GroupId (see Child's doc comment) and is
// immediately linked to the caller via a ParentChild row, exactly the same
// linking mechanism a placowka-connected child uses. Deliberately skips
// the ChildrenLimit paid-capacity check in Create's ValidateCreateMutation
// entirely (bypassed via gormdb.Create directly, not the staff Create
// flow) -- a personal child was never part of any organization's paid
// seat count.
func CreateMine(ctx context.Context, argument childrenpg.CreateMineArgument) (res childrenpg.CreateMineResponse, err error) {
	ctx = logger.Enter(ctx, "children.CreateMine", argument)
	defer logger.Exit(ctx, "children.CreateMine", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	child := childrenpg.Child{
		Id:         uuid.NewString(),
		FirstName:  argument.FirstName,
		LastName:   argument.LastName,
		BirthDate:  argument.BirthDate,
		CreateTime: time.Now(),
	}

	if err = gormdb.Create(ctx, &child); err != nil {
		return res, err
	}

	if _, err = parentchild.Create(ctx, parentchildpg.CreateArgument{
		ParentId: argument.ParentAccountId,
		ChildId:  child.Id,
		Relation: parentchildpg.RelationTypeGuardian,
	}); err != nil {
		return res, err
	}

	return childrenpg.CreateMineResponse{Child: child}, nil
}

// SelectMine returns every child the caller is linked to as a parent
// (ParentChild), whether personal (OrganizationId == "") or connected to a
// placowka -- the single, unified "my children" list the parent-facing
// mobile navigation is built around.
func SelectMine(ctx context.Context, parentAccountId string) (res childrenpg.SelectMineResponse, err error) {
	ctx = logger.Enter(ctx, "children.SelectMine", parentAccountId)
	defer logger.Exit(ctx, "children.SelectMine", &res, &err)

	links, err := parentchild.Select(ctx, parentchildpg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "parent_id", Operator: "=", Value: parentAccountId}}},
			},
		},
	})
	if err != nil {
		return res, err
	}

	if len(links.ParentChildren) == 0 {
		return childrenpg.SelectMineResponse{Children: []childrenpg.Child{}}, nil
	}

	childIds := make([]string, 0, len(links.ParentChildren))
	for _, link := range links.ParentChildren {
		childIds = append(childIds, link.ChildId)
	}

	childrenResult, err := Select(ctx, childrenpg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "id", Operator: "IN", Value: childIds}}},
			},
		},
	})
	if err != nil {
		return res, err
	}

	return childrenpg.SelectMineResponse{Children: childrenResult.Children}, nil
}
