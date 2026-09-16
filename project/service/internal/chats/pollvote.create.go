package chats

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"gorm.io/gorm"
)

var ErrPollClosed = errors.New("this poll is closed")
var ErrOptionNotInPoll = errors.New("option does not belong to this poll")
var ErrAlreadyVotedForOption = errors.New("already voted for this option")

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

	vote := argument.Populate()

	// The delete-then-insert (single choice) or check-then-insert (multi
	// choice) sequence has to happen inside one DB transaction serialized
	// per (poll, account), otherwise two concurrent votes from the same
	// account race: for a single-choice poll, both could pass the delete
	// and both insert, leaving two rows; for a multi-choice poll, both
	// could see "no existing vote for this option" and both insert,
	// letting one account vote for the same option any number of times
	// and inflate that option's tally. pg_advisory_xact_lock keyed on
	// poll_id+account_id (via hashtext) makes every concurrent call for
	// this account+poll queue up one at a time, mirroring the pattern
	// already used for the per-account channel limit in
	// chatinvites.Create.
	err = gormdb.GormDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", argument.PollId+":"+argument.AccountId).Error; err != nil {
			return err
		}

		if !poll.AllowMultiple {
			if err := tx.Where("poll_id = ? AND account_id = ?", argument.PollId, argument.AccountId).
				Delete(&chatspg.PollVote{}).Error; err != nil {
				return err
			}
		} else {
			var existing int64
			if err := tx.Model(&chatspg.PollVote{}).
				Where("poll_id = ? AND account_id = ? AND option_id = ?", argument.PollId, argument.AccountId, argument.OptionId).
				Count(&existing).Error; err != nil {
				return err
			}
			if existing > 0 {
				return ErrAlreadyVotedForOption
			}
		}

		return tx.Create(&vote).Error
	})
	if err != nil {
		return res, err
	}

	return chatspg.PollVoteCreateResponse{PollVote: vote}, nil
}
