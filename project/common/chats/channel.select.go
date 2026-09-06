package chatspg

import (
	"errors"
	"fmt"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/library/mutations"
	"github.com/go-playground/validator/v10"
)

type ChannelSelectArgument struct {
	Query databasepg.Query `json:"query,omitempty" validate:"required"`
}

func (a ChannelSelectArgument) Validate() error {
	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}
	return nil
}

type ChannelSelectResponse struct {
	Channels []Channel `json:"channels,omitempty"`
}

type ChannelSelectMutation mutations.Transitional[ChannelSelectArgument, []Channel]
