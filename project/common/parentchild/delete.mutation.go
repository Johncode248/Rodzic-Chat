package parentchildpg

import "github.com/Johncode248/Rodzic-Chat/library/mutations"

type DeleteMutation mutations.Conditional[DeleteArgument, DeleteCondition, []Transient]
