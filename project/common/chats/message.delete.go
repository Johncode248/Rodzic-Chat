package chatspg

import (
	"errors"
	"fmt"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/library/mutations"
	"github.com/go-playground/validator/v10"
)

type MessageDeleteArgument struct {
	Query databasepg.Query `json:"query,omitempty" validate:"required"`
}

func (a MessageDeleteArgument) Validate() error {
	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}
	return nil
}

type MessageDeleteResponse struct {
	Messages []Message `json:"messages,omitempty"`
}

// MessageDeleteCondition is empty: deleting a message does not cascade to
// its reads/reactions/poll rows in this version (future work).
type MessageDeleteCondition struct{}

type MessageDeleteMutation mutations.Conditional[MessageDeleteArgument, MessageDeleteCondition, []MessageTransient]
