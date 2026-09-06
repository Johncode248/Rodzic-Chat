package chatspg

import (
	"errors"
	"fmt"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/library/mutations"
	"github.com/go-playground/validator/v10"
)

type PollVoteDeleteArgument struct {
	Query databasepg.Query `json:"query,omitempty" validate:"required"`
}

func (a PollVoteDeleteArgument) Validate() error {
	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}
	return nil
}

type PollVoteDeleteRequest struct {
	PollId   string `json:"poll_id,omitempty"`
	OptionId string `json:"option_id,omitempty"`
}

type PollVoteDeleteResponse struct {
	PollVotes []PollVote `json:"poll_votes,omitempty"`
}

// PollVoteDeleteCondition is empty: retracting a vote has no cascading
// dependent deletes.
type PollVoteDeleteCondition struct{}

type PollVoteDeleteMutation mutations.Conditional[PollVoteDeleteArgument, PollVoteDeleteCondition, []PollVoteTransient]
