package chats

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
)

// DeleteParentChannelMessage deletes a message in an org-less parent chat: its author may always
// delete it; anyone else needs PermDeleteMessages AND must outrank the author (nobody deletes the
// admin's messages but the admin).
func DeleteParentChannelMessage(ctx context.Context, message chatspg.Message, callerAccountId string) (err error) {
	ctx = logger.Enter(ctx, "chats.DeleteParentChannelMessage", message.Id, callerAccountId)
	defer logger.Exit(ctx, "chats.DeleteParentChannelMessage", &err)

	if message.SenderId != callerAccountId {
		author, memberErr := getMember(ctx, message.ChannelId, message.SenderId)
		if memberErr != nil {
			return memberErr
		}
		if author != nil { // author already left -> only the permission matters
			if err = requireModeration(ctx, message.ChannelId, callerAccountId, author, chatspg.PermDeleteMessages); err != nil {
				return err
			}
		} else if ok, permErr := HasChannelPermission(ctx, message.ChannelId, callerAccountId, chatspg.PermDeleteMessages); permErr != nil || !ok {
			if permErr != nil {
				return permErr
			}
			return ErrNoPermission
		}
	}
	_, err = gormdb.Delete[chatspg.Message](ctx, databasepg.Query{Groups: []databasepg.Group{{Filters: []databasepg.Filter{
		{Key: "id", Operator: "=", Value: message.Id},
	}}}})
	return err
}
