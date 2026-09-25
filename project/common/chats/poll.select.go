package chatspg

import (
	"errors"
	"fmt"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/go-playground/validator/v10"
)

type PollSelectArgument struct {
	Query databasepg.Query `json:"query,omitempty" validate:"required"`
}

func (a PollSelectArgument) Validate() error {
	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}
	return nil
}

type PollSelectResponse struct {
	Polls []Poll `json:"polls,omitempty"`
}

type PollOptionSelectArgument struct {
	Query databasepg.Query `json:"query,omitempty" validate:"required"`
}

func (a PollOptionSelectArgument) Validate() error {
	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}
	return nil
}

type PollOptionSelectResponse struct {
	Options []PollOption `json:"options,omitempty"`
}

// PollOptionWithCount is one option plus its current vote tally, used by
// the poll detail aggregate response.
type PollOptionWithCount struct {
	Option    PollOption `json:"option,omitempty"`
	VoteCount int        `json:"vote_count,omitempty"`
}

// PollDetailResponse is the aggregate "fetch a poll" response: the poll,
// its options with tallies, and the caller's own vote(s) -- not raw CRUD.
type PollDetailResponse struct {
	Poll    Poll                  `json:"poll,omitempty"`
	Options []PollOptionWithCount `json:"options,omitempty"`
	MyVotes []string              `json:"my_votes,omitempty"`
	// Voters maps option id -> display names of who voted for it. Empty for an anonymous poll.
	Voters map[string][]string `json:"voters,omitempty"`
}
