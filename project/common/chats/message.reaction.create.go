package chatspg

import (
	"errors"
	"fmt"
	"time"

	"github.com/Johncode248/Rodzic-Chat/library/expressions"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type MessageReactionCreateArgument struct {
	Id        string `json:"id,omitempty"`
	MessageId string `json:"message_id,omitempty" validate:"required"`
	AccountId string `json:"account_id,omitempty" validate:"required"`
	Emoji     string `json:"emoji,omitempty" validate:"required,max=32"`
}

func (a MessageReactionCreateArgument) Validate() error {
	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}
	return nil
}

func (a MessageReactionCreateArgument) Populate() (reaction MessageReaction) {
	reaction.Id = expressions.Select(a.Id != "", a.Id, uuid.NewString())
	reaction.MessageId = a.MessageId
	reaction.AccountId = a.AccountId
	reaction.Emoji = a.Emoji
	reaction.CreateTime = time.Now()
	return reaction
}

type MessageReactionCreateRequest struct {
	MessageId string `json:"message_id,omitempty"`
	Emoji     string `json:"emoji,omitempty"`
}

type MessageReactionCreateResponse struct {
	MessageReaction MessageReaction `json:"message_reaction,omitempty"`
}
