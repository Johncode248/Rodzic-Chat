package chats

import (
	"context"
	"errors"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
)

func SelectPolls(ctx context.Context, argument chatspg.PollSelectArgument) (res chatspg.PollSelectResponse, err error) {
	ctx = logger.Enter(ctx, "chats.SelectPolls", argument)
	defer logger.Exit(ctx, "chats.SelectPolls", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	result, err := gormdb.Select[chatspg.Poll](ctx, argument.Query)
	if err != nil {
		return res, err
	}

	return chatspg.PollSelectResponse{Polls: result}, nil
}

func SelectPollOptions(ctx context.Context, argument chatspg.PollOptionSelectArgument) (res chatspg.PollOptionSelectResponse, err error) {
	ctx = logger.Enter(ctx, "chats.SelectPollOptions", argument)
	defer logger.Exit(ctx, "chats.SelectPollOptions", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	result, err := gormdb.Select[chatspg.PollOption](ctx, argument.Query)
	if err != nil {
		return res, err
	}

	return chatspg.PollOptionSelectResponse{Options: result}, nil
}

func SelectPollVotes(ctx context.Context, argument chatspg.PollVoteSelectArgument) (res chatspg.PollVoteSelectResponse, err error) {
	ctx = logger.Enter(ctx, "chats.SelectPollVotes", argument)
	defer logger.Exit(ctx, "chats.SelectPollVotes", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	result, err := gormdb.Select[chatspg.PollVote](ctx, argument.Query)
	if err != nil {
		return res, err
	}

	return chatspg.PollVoteSelectResponse{PollVotes: result}, nil
}

// GetPollDetail builds the poll+options+tally+"my votes" aggregate for
// GET /private/chats/polls?poll_id=.
func GetPollDetail(ctx context.Context, pollId, callerAccountId string) (res chatspg.PollDetailResponse, err error) {
	ctx = logger.Enter(ctx, "chats.GetPollDetail", pollId)
	defer logger.Exit(ctx, "chats.GetPollDetail", &res, &err)

	pollResult, err := SelectPolls(ctx, chatspg.PollSelectArgument{
		Query: databasepg.Query{
			Limit:  1,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: pollId}}}},
		},
	})
	if err != nil {
		return res, err
	}
	if len(pollResult.Polls) == 0 {
		return res, errors.New("poll not found")
	}
	poll := pollResult.Polls[0]

	optionsResult, err := SelectPollOptions(ctx, chatspg.PollOptionSelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "poll_id", Operator: "=", Value: pollId}}}},
		},
	})
	if err != nil {
		return res, err
	}

	votesResult, err := SelectPollVotes(ctx, chatspg.PollVoteSelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "poll_id", Operator: "=", Value: pollId}}}},
		},
	})
	if err != nil {
		return res, err
	}

	counts := make(map[string]int, len(optionsResult.Options))
	var myVotes []string
	for _, vote := range votesResult.PollVotes {
		counts[vote.OptionId]++
		if vote.AccountId == callerAccountId {
			myVotes = append(myVotes, vote.OptionId)
		}
	}

	optionsWithCount := make([]chatspg.PollOptionWithCount, 0, len(optionsResult.Options))
	for _, option := range optionsResult.Options {
		optionsWithCount = append(optionsWithCount, chatspg.PollOptionWithCount{
			Option:    option,
			VoteCount: counts[option.Id],
		})
	}

	return chatspg.PollDetailResponse{
		Poll:    poll,
		Options: optionsWithCount,
		MyVotes: myVotes,
	}, nil
}
