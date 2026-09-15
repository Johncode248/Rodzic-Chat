package chats

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
)

// AttachSenderNames sets Message.SenderName in place for each message, batch-resolving every
// distinct SenderId to its account's display name in one query -- the chat UI shows this instead
// of the raw account id (see MessageBubble on both clients). A blank SenderId (system message)
// or an id that no longer resolves to an account is left as "".
func AttachSenderNames(ctx context.Context, messages []chatspg.Message) ([]chatspg.Message, error) {
	senderIds := make([]string, 0, len(messages))
	seen := map[string]bool{}
	for _, m := range messages {
		if m.SenderId != "" && !seen[m.SenderId] {
			seen[m.SenderId] = true
			senderIds = append(senderIds, m.SenderId)
		}
	}
	if len(senderIds) == 0 {
		return messages, nil
	}

	accountResult, err := accounts.Select(ctx, accountspg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "IN", Value: senderIds}}}},
		},
	})
	if err != nil {
		return nil, err
	}
	nameById := map[string]string{}
	for _, a := range accountResult.Accounts {
		nameById[a.Id] = a.Name
	}

	for i, m := range messages {
		messages[i].SenderName = nameById[m.SenderId]
	}
	return messages, nil
}
