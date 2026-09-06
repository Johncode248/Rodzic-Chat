package chatspg

import (
	"errors"
	"fmt"
	"time"

	"github.com/Johncode248/Rodzic-Chat/library/expressions"
	"github.com/Johncode248/Rodzic-Chat/library/mutations"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type PollVoteCreateArgument struct {
	Id        string `json:"id,omitempty"`
	PollId    string `json:"poll_id,omitempty" validate:"required"`
	OptionId  string `json:"option_id,omitempty" validate:"required"`
	AccountId string `json:"account_id,omitempty" validate:"required"`
}

func (a PollVoteCreateArgument) Validate() error {
	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}
	return nil
}

func (a PollVoteCreateArgument) Populate() (vote PollVote) {
	vote.Id = expressions.Select(a.Id != "", a.Id, uuid.NewString())
	vote.PollId = a.PollId
	vote.OptionId = a.OptionId
	vote.AccountId = a.AccountId
	vote.CreateTime = time.Now()
	return vote
}

type PollVoteCreateRequest struct {
	PollId   string `json:"poll_id,omitempty"`
	OptionId string `json:"option_id,omitempty"`
}

type PollVoteCreateResponse struct {
	PollVote PollVote `json:"poll_vote,omitempty"`
}

type PollVoteCreateMutation mutations.Transitional[PollVoteCreateArgument, PollVoteTransient]
