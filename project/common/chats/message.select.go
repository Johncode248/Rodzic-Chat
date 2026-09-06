package chatspg

import (
	"errors"
	"fmt"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/library/mutations"
	"github.com/go-playground/validator/v10"
)

type MessageSelectArgument struct {
	Query databasepg.Query `json:"query,omitempty" validate:"required"`
}

func (a MessageSelectArgument) Validate() error {
	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}
	return nil
}

type MessageSelectResponse struct {
	Messages []Message `json:"messages,omitempty"`
}

type MessageSelectMutation mutations.Transitional[MessageSelectArgument, []Message]
