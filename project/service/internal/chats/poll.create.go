package chats

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/google/uuid"
)

// CreatePoll creates a Message (Type=Poll) + Poll + PollOptions together,
// applying the same channel-type posting-permission and rate-limit checks
// as a normal text message (a poll is a message). Steps are sequential
// since PollOptions need the Poll's id which needs the Message's id.
func CreatePoll(ctx context.Context, argument chatspg.PollCreateArgument, accountType int) (res chatspg.PollCreateResponse, err error) {
	ctx = logger.Enter(ctx, "chats.CreatePoll", argument)
	defer logger.Exit(ctx, "chats.CreatePoll", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	messageResult, err := CreateMessage(ctx, chatspg.MessageCreateArgument{
		ChannelId:      argument.ChannelId,
		OrganizationId: argument.OrganizationId,
		SenderId:       argument.SenderId,
		Type:           chatspg.MessageTypePoll,
	}, accountType)
	if err != nil {
		return res, err
	}

	message := messageResult.Message

	pollRow := chatspg.PollRowCreateArgument{
		MessageId:     message.Id,
		Question:      argument.Question,
		AllowMultiple: argument.AllowMultiple,
		ClosesAt:      argument.ClosesAt,
	}.Populate()

	if err = gormdb.Create(ctx, pollRow); err != nil {
		_ = FinalizeMessageDeleteMutation(ctx, &chatspg.MessageDeleteMutation{
			Foundation: chatspg.MessageDeleteArgument{
				Query: databasepg.Query{
					Groups: []databasepg.Group{
						{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: message.Id}}},
					},
				},
			},
		})
		return res, err
	}

	options := make([]chatspg.PollOption, 0, len(argument.Options))
	for index, text := range argument.Options {
		option := chatspg.PollOptionCreateArgument{
			Id:         uuid.NewString(),
			PollId:     pollRow.Id,
			Text:       text,
			OrderIndex: index,
		}.Populate()

		if err = gormdb.Create(ctx, option); err != nil {
			return res, err
		}

		options = append(options, option)
	}

	return chatspg.PollCreateResponse{
		Message: message,
		Poll:    pollRow,
		Options: options,
	}, nil
}
