package databasepg

import (
	"fmt"

	"github.com/supabase-community/postgrest-go"
)

type Filter struct {
	Key      string `json:"key,omitempty"`
	Operator string `json:"operator,omitempty"`
	Value    any    `json:"value,omitempty"`
}

func ApplyFilter(query *postgrest.FilterBuilder, cond Filter) *postgrest.FilterBuilder {
	switch cond.Operator {
	case "eq":
		return query.Eq(cond.Key, cond.Value.(string))
	case "neq":
		return query.Neq(cond.Key, cond.Value.(string))
	case "gt":
		return query.Gt(cond.Key, cond.Value.(string))
	case "gte":
		return query.Gte(cond.Key, cond.Value.(string))
	case "lt":
		return query.Lt(cond.Key, cond.Value.(string))
	case "lte":
		return query.Lte(cond.Key, cond.Value.(string))
	case "like":
		return query.Like(cond.Key, fmt.Sprintf("%v", cond.Value))
	case "in":
		return query.In(cond.Key, cond.Value.([]string))
	default:
		return query.Eq(cond.Key, cond.Value.(string))
	}
}
