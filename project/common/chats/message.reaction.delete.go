package chatspg

import (
	"errors"
	"fmt"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/library/mutations"
	"github.com/go-playground/validator/v10"
)

type MessageReactionDeleteArgument struct {
	Query databasepg.Query `json:"query,omitempty" validate:"required"`
}

func (a MessageReactionDeleteArgument) Validate() error {
	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}
	return nil
}

type MessageReactionDeleteResponse struct {
	MessageReactions []MessageReaction `json:"message_reactions,omitempty"`
}

// MessageReactionDeleteCondition is empty: removing a reaction has no
// cascading dependent deletes.
type MessageReactionDeleteCondition struct{}

type MessageReactionDeleteMutation mutations.Conditional[MessageReactionDeleteArgument, MessageReactionDeleteCondition, []MessageReactionTransient]
