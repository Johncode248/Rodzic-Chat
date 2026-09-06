package eventspg

import "github.com/Johncode248/Rodzic-Chat/library/mutations"

type EventDeleteMutation mutations.Conditional[EventDeleteArgument, EventDeleteCondition, []EventTransient]
