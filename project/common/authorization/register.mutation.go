package authorizationpg

import (
	"github.com/Johncode248/Rodzic-Chat/library/mutations"
)

type RegisterMutation mutations.Conditional[RegisterArgument, RegisterCondition, string]
