package chats

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrNotSignupMessage = errors.New("message is not a signup list")
	ErrSlotFull         = errors.New("Wszystkie miejsca w tej pozycji są już zajęte")
	ErrSlotNotFound     = errors.New("signup slot not found")
)

const maxSignupSlots = 30

// CreateSignup posts a Signup-type message plus its slots. Anyone allowed to post in the channel
// may create one (a teacher's "kto przyniesie ciasto", a parent's list in the parents channel).
func CreateSignup(ctx context.Context, channelId, organizationId, senderId string, request chatspg.SignupCreateRequest, accountType int) (res chatspg.SignupCreateResponse, err error) {
	ctx = logger.Enter(ctx, "chats.CreateSignup", channelId)
	defer logger.Exit(ctx, "chats.CreateSignup", &res, &err)

	if request.Title == "" {
		return res, errors.New("signup title is required")
	}
	if len(request.Slots) == 0 || len(request.Slots) > maxSignupSlots {
		return res, errors.New("a signup list needs between 1 and 30 items")
	}
	for _, slot := range request.Slots {
		if slot.Title == "" || slot.Capacity < 0 {
			return res, errors.New("every signup item needs a title and a non-negative capacity")
		}
	}

	messageResult, err := CreateMessage(ctx, chatspg.MessageCreateArgument{
		ChannelId:      channelId,
		OrganizationId: organizationId,
		SenderId:       senderId,
		Type:           chatspg.MessageTypeSignup,
		Content:        request.Title,
	}, accountType)
	if err != nil {
		return res, err
	}

	for index, input := range request.Slots {
		slot := chatspg.SignupSlot{
			Id:         uuid.NewString(),
			MessageId:  messageResult.Message.Id,
			Title:      input.Title,
			Capacity:   input.Capacity,
			OrderIndex: index,
		}
		if err = gormdb.GormDB.WithContext(ctx).Create(&slot).Error; err != nil {
			return res, err
		}
	}

	detail, err := GetSignupDetail(ctx, messageResult.Message.Id, senderId)
	if err != nil {
		return res, err
	}
	return chatspg.SignupCreateResponse{Message: messageResult.Message, Detail: detail}, nil
}

// GetSignupDetail returns the list with, per slot, who claimed it and whether the caller did.
func GetSignupDetail(ctx context.Context, messageId, callerId string) (res chatspg.SignupDetailResponse, err error) {
	ctx = logger.Enter(ctx, "chats.GetSignupDetail", messageId)
	defer logger.Exit(ctx, "chats.GetSignupDetail", &res, &err)

	message, err := LoadMessage(ctx, messageId)
	if err != nil {
		return res, err
	}
	if message.Type != chatspg.MessageTypeSignup {
		return res, ErrNotSignupMessage
	}

	var slots []chatspg.SignupSlot
	if err = gormdb.GormDB.WithContext(ctx).Where("message_id = ?", messageId).Order("order_index asc").Find(&slots).Error; err != nil {
		return res, err
	}
	slotIds := make([]string, 0, len(slots))
	for _, s := range slots {
		slotIds = append(slotIds, s.Id)
	}
	var claims []chatspg.SignupClaim
	if len(slotIds) > 0 {
		if err = gormdb.GormDB.WithContext(ctx).Where("slot_id IN ?", slotIds).Order("create_time asc").Find(&claims).Error; err != nil {
			return res, err
		}
	}
	accountIds := make([]string, 0, len(claims))
	for _, c := range claims {
		accountIds = append(accountIds, c.AccountId)
	}
	names, err := AccountNames(ctx, accountIds)
	if err != nil {
		return res, err
	}

	claimsBySlot := map[string][]chatspg.SignupClaim{}
	for _, c := range claims {
		claimsBySlot[c.SlotId] = append(claimsBySlot[c.SlotId], c)
	}
	res.MessageId = messageId
	res.Title = message.Content
	for _, slot := range slots {
		detail := chatspg.SignupSlotDetail{Slot: slot}
		for _, c := range claimsBySlot[slot.Id] {
			at := c.CreateTime
			detail.Claimed = append(detail.Claimed, chatspg.PersonRef{AccountId: c.AccountId, Name: names[c.AccountId], Time: &at})
			if c.AccountId == callerId {
				detail.IClaimed = true
			}
		}
		detail.Full = slot.Capacity > 0 && len(detail.Claimed) >= slot.Capacity
		res.Slots = append(res.Slots, detail)
	}
	sort.SliceStable(res.Slots, func(i, j int) bool { return res.Slots[i].Slot.OrderIndex < res.Slots[j].Slot.OrderIndex })
	return res, nil
}

// SlotMessageId resolves a slot to the message that owns it (for authorization).
func SlotMessageId(ctx context.Context, slotId string) (string, error) {
	var slot chatspg.SignupSlot
	err := gormdb.GormDB.WithContext(ctx).Where("id = ?", slotId).First(&slot).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", ErrSlotNotFound
	}
	return slot.MessageId, err
}

// ClaimSlot takes a spot in slotId for accountId. The slot row is locked for the duration so two
// people racing for the last place can't both get it.
func ClaimSlot(ctx context.Context, slotId, accountId string) (err error) {
	ctx = logger.Enter(ctx, "chats.ClaimSlot", slotId)
	defer logger.Exit(ctx, "chats.ClaimSlot", &err)

	return gormdb.GormDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var slot chatspg.SignupSlot
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", slotId).First(&slot).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrSlotNotFound
			}
			return err
		}
		var existing int64
		if err := tx.Model(&chatspg.SignupClaim{}).Where("slot_id = ? AND account_id = ?", slotId, accountId).Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			return nil
		}
		if slot.Capacity > 0 {
			var taken int64
			if err := tx.Model(&chatspg.SignupClaim{}).Where("slot_id = ?", slotId).Count(&taken).Error; err != nil {
				return err
			}
			if int(taken) >= slot.Capacity {
				return ErrSlotFull
			}
		}
		claim := chatspg.SignupClaim{Id: uuid.NewString(), SlotId: slotId, AccountId: accountId, CreateTime: time.Now()}
		return tx.Create(&claim).Error
	})
}

// UnclaimSlot gives up accountId's spot in slotId (a no-op if they had none).
func UnclaimSlot(ctx context.Context, slotId, accountId string) (err error) {
	ctx = logger.Enter(ctx, "chats.UnclaimSlot", slotId)
	defer logger.Exit(ctx, "chats.UnclaimSlot", &err)

	return gormdb.GormDB.WithContext(ctx).Where("slot_id = ? AND account_id = ?", slotId, accountId).Delete(&chatspg.SignupClaim{}).Error
}
