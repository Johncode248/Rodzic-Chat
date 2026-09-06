package chats

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
)

func DeleteMessage(ctx context.Context, argument chatspg.MessageDeleteArgument) (res chatspg.MessageDeleteResponse, err error) {
	ctx = logger.Enter(ctx, "chats.DeleteMessage", argument)
	defer logger.Exit(ctx, "chats.DeleteMessage", &res, &err)

	var (
		usesMutation = &chatspg.MessageDeleteMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = InitiateMessageDeleteMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	defer func() {
		if err != nil {
			if err = RollbackMessageDeleteMutation(ctx, usesMutation); err != nil {
				panic(err)
			}
		}
	}()

	if err = FinalizeMessageDeleteMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	for _, value := range usesMutation.Transient {
		res.Messages = append(res.Messages, value.Previous)
	}

	return res, nil
}

func InitiateMessageDeleteMutation(ctx context.Context, mutation *chatspg.MessageDeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "chats.InitiateMessageDeleteMutation", mutation)
	defer logger.Exit(ctx, "chats.InitiateMessageDeleteMutation", &err)

	result, err := SelectMessages(ctx, chatspg.MessageSelectArgument{Query: mutation.Foundation.Query})
	if err != nil {
		return err
	}

	for _, message := range result.Messages {
		mutation.Transient = append(mutation.Transient, chatspg.MessageTransient{Current: message})
	}

	return nil
}

func FinalizeMessageDeleteMutation(ctx context.Context, mutation *chatspg.MessageDeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "chats.FinalizeMessageDeleteMutation", mutation)
	defer logger.Exit(ctx, "chats.FinalizeMessageDeleteMutation", &err)

	if _, err = gormdb.Delete[chatspg.Message](ctx, mutation.Foundation.Query); err != nil {
		return err
	}

	for index := range mutation.Transient {
		mutation.Transient[index] = mutation.Transient[index].Update()
	}

	return nil
}

func RollbackMessageDeleteMutation(ctx context.Context, mutation *chatspg.MessageDeleteMutation) (err error) {
	ctx = logger.Enter(ctx, "chats.RollbackMessageDeleteMutation", mutation)
	defer logger.Exit(ctx, "chats.RollbackMessageDeleteMutation", &err)

	return routines.Select(
		func() error {
			functions := []func() error{}
			for _, transient := range mutation.Transient {
				if transient.Previous.Id == "" {
					continue
				}
				transient := transient
				functions = append(functions, func() error {
					usesMutation := &chatspg.MessageCreateMutation{
						Transient: chatspg.MessageTransient{Future: transient.Previous},
					}
					return FinalizeMessageCreateMutation(ctx, usesMutation)
				})
			}
			return routines.Select(functions...)
		},
	)
}
