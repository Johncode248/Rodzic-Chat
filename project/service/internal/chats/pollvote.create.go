package chats

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
)

var ErrPollClosed = errors.New("this poll is closed")
var ErrOptionNotInPoll = errors.New("option does not belong to this poll")

// CreateVote casts (or, for a single-choice poll, changes) a vote: it
// validates the poll isn't closed and the option belongs to the poll, and
// when AllowMultiple is false it deletes the account's prior vote(s) on
// this poll first so re-voting replaces rather than adds a second vote.
func CreateVote(ctx context.Context, argument chatspg.PollVoteCreateArgument) (res chatspg.PollVoteCreateResponse, err error) {
	ctx = logger.Enter(ctx, "chats.CreateVote", argument)
	defer logger.Exit(ctx, "chats.CreateVote", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	pollResult, err := SelectPolls(ctx, chatspg.PollSelectArgument{
		Query: databasepg.Query{
			Limit:  1,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: argument.PollId}}}},
		},
	})
	if err != nil {
		return res, err
	}
	if len(pollResult.Polls) == 0 {
		return res, errors.New("poll not found")
	}
	poll := pollResult.Polls[0]

	if poll.ClosesAt != nil && poll.ClosesAt.Before(time.Now()) {
		return res, ErrPollClosed
	}

	optionResult, err := SelectPollOptions(ctx, chatspg.PollOptionSelectArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{
					{Key: "id", Operator: "=", Value: argument.OptionId},
					{Key: "poll_id", Operator: "=", Value: argument.PollId},
				}},
			},
		},
	})
	if err != nil {
		return res, err
	}
	if len(optionResult.Options) == 0 {
		return res, ErrOptionNotInPoll
	}

	if !poll.AllowMultiple {
		if _, err = gormdb.Delete[chatspg.PollVote](ctx, databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{
					{Key: "poll_id", Operator: "=", Value: argument.PollId},
					{Key: "account_id", Operator: "=", Value: argument.AccountId},
				}},
			},
		}); err != nil {
			return res, err
		}
	}

	vote := argument.Populate()
	if err = gormdb.Create(ctx, vote); err != nil {
		return res, err
	}

	return chatspg.PollVoteCreateResponse{PollVote: vote}, nil
}
