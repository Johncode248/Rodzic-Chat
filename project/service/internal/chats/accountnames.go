package chats

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
)

// AccountNames batch-resolves account ids to display names (one query). Unknown ids are absent.
func AccountNames(ctx context.Context, ids []string) (map[string]string, error) {
	names := map[string]string{}
	unique := make([]string, 0, len(ids))
	seen := map[string]bool{}
	for _, id := range ids {
		if id != "" && !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}
	if len(unique) == 0 {
		return names, nil
	}
	result, err := accounts.Select(ctx, accountspg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "IN", Value: unique}}}},
		},
	})
	if err != nil {
		return nil, err
	}
	for _, a := range result.Accounts {
		names[a.Id] = a.Name
	}
	return names, nil
}

// accountTypes is like AccountNames but returns account types (for splitting staff/parents).
func accountTypes(ctx context.Context, ids []string) (map[string]int, error) {
	types := map[string]int{}
	if len(ids) == 0 {
		return types, nil
	}
	result, err := accounts.Select(ctx, accountspg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "IN", Value: ids}}}},
		},
	})
	if err != nil {
		return nil, err
	}
	for _, a := range result.Accounts {
		types[a.Id] = int(a.AccountType)
	}
	return types, nil
}
