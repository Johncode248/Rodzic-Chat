package chatspg

import (
	"errors"
	"fmt"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/library/mutations"
	"github.com/go-playground/validator/v10"
)

type ChannelDeleteArgument struct {
	Query databasepg.Query `json:"query,omitempty" validate:"required"`
}

func (a ChannelDeleteArgument) Validate() error {
	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}
	return nil
}

type ChannelDeleteResponse struct {
	Channels []Channel `json:"channels,omitempty"`
}

// ChannelDeleteCondition is empty: deleting a channel does not cascade to
// its messages in this version (future work).
type ChannelDeleteCondition struct{}

type ChannelDeleteMutation mutations.Conditional[ChannelDeleteArgument, ChannelDeleteCondition, []ChannelTransient]
