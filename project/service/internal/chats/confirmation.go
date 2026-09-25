package chats

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/push"
	"github.com/google/uuid"
	"gorm.io/gorm/clause"
)

var ErrNotConsentMessage = errors.New("message is not a consent request")

// CreateConsent posts a Consent-type message ("please confirm you've read/agree"). Who may post
// into the channel and the rate limit are enforced by CreateMessage, like any other message.
func CreateConsent(ctx context.Context, channelId, organizationId, senderId, content string, accountType int) (res chatspg.ConsentCreateResponse, err error) {
	ctx = logger.Enter(ctx, "chats.CreateConsent", channelId)
	defer logger.Exit(ctx, "chats.CreateConsent", &res, &err)

	if content == "" {
		return res, errors.New("consent text is required")
	}
	messageResult, err := CreateMessage(ctx, chatspg.MessageCreateArgument{
		ChannelId:      channelId,
		OrganizationId: organizationId,
		SenderId:       senderId,
		Type:           chatspg.MessageTypeConsent,
		Content:        content,
	}, accountType)
	if err != nil {
		return res, err
	}
	return chatspg.ConsentCreateResponse{Message: messageResult.Message}, nil
}

// ConfirmConsent records that accountId confirmed messageId (idempotent).
func ConfirmConsent(ctx context.Context, messageId, accountId string) (err error) {
	ctx = logger.Enter(ctx, "chats.ConfirmConsent", messageId)
	defer logger.Exit(ctx, "chats.ConfirmConsent", &err)

	message, err := LoadMessage(ctx, messageId)
	if err != nil {
		return err
	}
	if message.Type != chatspg.MessageTypeConsent {
		return ErrNotConsentMessage
	}
	confirmation := chatspg.MessageConfirmation{
		Id:          uuid.NewString(),
		MessageId:   messageId,
		AccountId:   accountId,
		ConfirmTime: time.Now(),
	}
	return gormdb.GormDB.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&confirmation).Error
}

// consentParents returns the parent accounts that can read channel -- the people a consent is
// actually asked of (staff can read every channel but are never expected to confirm).
func consentParents(ctx context.Context, channel chatspg.Channel) ([]string, error) {
	recipients, err := Recipients(ctx, channel)
	if err != nil {
		return nil, err
	}
	types, err := accountTypes(ctx, recipients)
	if err != nil {
		return nil, err
	}
	parents := make([]string, 0, len(recipients))
	for _, id := range recipients {
		if types[id] == accountspg.AccountTypeParent.Ordinal() {
			parents = append(parents, id)
		}
	}
	return parents, nil
}

// ConsentStatus builds the caller's view of a consent. Staff additionally get the named
// confirmed/pending lists.
func ConsentStatus(ctx context.Context, message chatspg.Message, channel chatspg.Channel, callerId string, callerIsStaff bool) (res chatspg.ConsentStatusResponse, err error) {
	ctx = logger.Enter(ctx, "chats.ConsentStatus", message.Id)
	defer logger.Exit(ctx, "chats.ConsentStatus", &res, &err)

	if message.Type != chatspg.MessageTypeConsent {
		return res, ErrNotConsentMessage
	}

	var confirmations []chatspg.MessageConfirmation
	if err = gormdb.GormDB.WithContext(ctx).Where("message_id = ?", message.Id).Order("confirm_time asc").Find(&confirmations).Error; err != nil {
		return res, err
	}
	parents, err := consentParents(ctx, channel)
	if err != nil {
		return res, err
	}

	confirmedIds := map[string]time.Time{}
	for _, c := range confirmations {
		confirmedIds[c.AccountId] = c.ConfirmTime
	}
	_, res.IConfirmed = confirmedIds[callerId]
	res.MessageId = message.Id
	res.TotalCount = len(parents)
	for _, id := range parents {
		if _, ok := confirmedIds[id]; ok {
			res.ConfirmedCount++
		}
	}
	// A confirmation from someone no longer in the recipient list still counts toward the total.
	if res.ConfirmedCount > res.TotalCount {
		res.TotalCount = res.ConfirmedCount
	}

	if !callerIsStaff {
		return res, nil
	}

	allIds := append([]string{}, parents...)
	for id := range confirmedIds {
		allIds = append(allIds, id)
	}
	names, err := AccountNames(ctx, allIds)
	if err != nil {
		return res, err
	}
	for id, at := range confirmedIds {
		t := at
		res.Confirmed = append(res.Confirmed, chatspg.PersonRef{AccountId: id, Name: names[id], Time: &t})
	}
	sort.Slice(res.Confirmed, func(i, j int) bool { return res.Confirmed[i].Time.Before(*res.Confirmed[j].Time) })
	for _, id := range parents {
		if _, ok := confirmedIds[id]; !ok {
			res.Pending = append(res.Pending, chatspg.PersonRef{AccountId: id, Name: names[id]})
		}
	}
	sort.Slice(res.Pending, func(i, j int) bool { return res.Pending[i].Name < res.Pending[j].Name })
	return res, nil
}

// RemindPending pushes a reminder to every parent who has not confirmed yet and returns how many
// were reminded.
func RemindPending(ctx context.Context, message chatspg.Message, channel chatspg.Channel) (count int, err error) {
	ctx = logger.Enter(ctx, "chats.RemindPending", message.Id)
	defer logger.Exit(ctx, "chats.RemindPending", &count, &err)

	status, err := ConsentStatus(ctx, message, channel, "", true)
	if err != nil {
		return 0, err
	}
	targets := make([]string, 0, len(status.Pending))
	for _, p := range status.Pending {
		targets = append(targets, p.AccountId)
	}
	if len(targets) == 0 {
		return 0, nil
	}
	body := message.Content
	if len(body) > 120 {
		body = body[:120] + "…"
	}
	push.SendToAccounts(ctx, targets, "Przypomnienie: potwierdź", body, map[string]string{
		"type":       "chat_message",
		"channel_id": channel.Id,
	})
	return len(targets), nil
}
