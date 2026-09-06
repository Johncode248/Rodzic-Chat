package chatspg

import (
	"errors"
	"fmt"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/go-playground/validator/v10"
)

type PollVoteSelectArgument struct {
	Query databasepg.Query `json:"query,omitempty" validate:"required"`
}

func (a PollVoteSelectArgument) Validate() error {
	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}
	return nil
}

type PollVoteSelectResponse struct {
	PollVotes []PollVote `json:"poll_votes,omitempty"`
}
