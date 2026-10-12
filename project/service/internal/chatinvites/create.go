package chatinvites

import (
	"context"
	"errors"
	"fmt"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/pinhash"
	"github.com/Johncode248/Rodzic-Chat/library/random"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	chatinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/chatinvites"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const expiryDuration = 72 * time.Hour

// maxPinGenerationAttempts bounds retrying GeneratePin on the rare
// uniqueIndex collision on pin_hash.
const maxPinGenerationAttempts = 5

// MaxChannelsPerAccount caps how many org-less parent channels one account
// may create (be the admin of) -- a parent can still join as many other
// people's channels as they like, this only limits how many they start
// themselves. A product judgment call to keep this feature from becoming
// an unbounded free group-chat host; revisit if it turns out too strict.
const MaxChannelsPerAccount = 3

var ErrChannelLimitReached = errors.New("możesz utworzyć maksymalnie 3 kanały -- opuść jeden, aby założyć nowy")

func resolveDisplayName(ctx context.Context, accountId string) string {
	if accountResult, accErr := accounts.Select(ctx, accountspg.SelectArgument{
		Query: databasepg.Query{Limit: 1, Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: accountId}}},
		}},
	}); accErr == nil && len(accountResult.Accounts) > 0 && accountResult.Accounts[0].Name != "" {
		return accountResult.Accounts[0].Name
	}
	return "Rodzic"
}

// generatePinHash generates a fresh plaintext PIN and its hash -- the
// caller is responsible for retrying the whole insert on a uniqueIndex
// collision (see Create/CreateForChannel below).
func generatePinHash() (plaintext string, hash string) {
	plaintext = random.GeneratePin()
	return plaintext, pinhash.HashPin(plaintext)
}

// Create sets up a brand-new, org-less channel for accountId (Type Direct,
// OrganizationId "") -- the "rodzice między sobą" case where parents share
// no placówka -- adds accountId as its Admin ChannelMember, and issues a
// PIN to share with the first invitee (see CreateForChannel to invite more
// people into an already-existing channel instead of starting a new one).
// Capped at [MaxChannelsPerAccount] channels created per account. Returns
// the plaintext PIN alongside the stored invite -- it's never retrievable
// again.
func Create(ctx context.Context, accountId string) (invite chatinvitespg.ChatInvite, channel chatspg.Channel, plaintextPin string, err error) {
	ctx = logger.Enter(ctx, "chatinvites.Create", accountId)
	defer logger.Exit(ctx, "chatinvites.Create", &invite, &channel, &err)

	now := time.Now()
	inviterName := resolveDisplayName(ctx, accountId)

	channel = chatspg.Channel{
		Id:         uuid.NewString(),
		Type:       chatspg.ChannelTypeDirect,
		Name:       fmt.Sprintf("Czat: %s", inviterName),
		CreateTime: now,
		UpdateTime: now,
	}
	member := chatspg.ChannelMember{
		Id:         uuid.NewString(),
		ChannelId:  channel.Id,
		AccountId:  accountId,
		Role:       chatspg.ChannelMemberRoleAdmin,
		CreateTime: now,
	}

	// The count-then-insert MaxChannelsPerAccount check has to happen
	// inside one DB transaction serialized per-account, otherwise N
	// concurrent Create calls for the same account can all read the same
	// (pre-limit) count and all insert, blowing past the cap -- a classic
	// TOCTOU race. pg_advisory_xact_lock(hashtext(accountId)) makes every
	// concurrent call for this account queue up one at a time (auto-released
	// when the transaction ends), so the re-count each one sees is always
	// up to date with every insert that committed before it.
	for attempt := 0; ; attempt++ {
		var pinHash string
		plaintextPin, pinHash = generatePinHash()
		invite = chatinvitespg.ChatInvite{
			Id:               uuid.NewString(),
			PinHash:          pinHash,
			ChannelId:        channel.Id,
			InviterAccountId: accountId,
			ExpiresAt:        now.Add(expiryDuration),
			CreateTime:       now,
		}

		err = gormdb.GormDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", accountId).Error; err != nil {
				return err
			}

			var adminCount int64
			if err := tx.Model(&chatspg.ChannelMember{}).
				Where("account_id = ? AND role = ?", accountId, int(chatspg.ChannelMemberRoleAdmin)).
				Count(&adminCount).Error; err != nil {
				return err
			}
			if adminCount >= MaxChannelsPerAccount {
				return ErrChannelLimitReached
			}

			if err := tx.Create(&channel).Error; err != nil {
				return err
			}
			if err := tx.Create(&member).Error; err != nil {
				return err
			}
			if err := tx.Create(&invite).Error; err != nil {
				return err
			}
			return nil
		})
		if err == nil {
			break
		}
		if !pinhash.IsUniqueViolation(err) || attempt >= maxPinGenerationAttempts-1 {
			return chatinvitespg.ChatInvite{}, chatspg.Channel{}, "", err
		}
	}

	return invite, channel, plaintextPin, nil
}

// CreateForChannel issues an additional invite PIN for an existing channel
// -- "Zaproś kolejną osobę" from within the channel's settings, once it
// already has at least one member beyond its creator. Only the channel's
// admin may do this. Returns the plaintext PIN alongside the stored
// invite.
func CreateForChannel(ctx context.Context, channelId, accountId string) (invite chatinvitespg.ChatInvite, plaintextPin string, err error) {
	ctx = logger.Enter(ctx, "chatinvites.CreateForChannel", channelId, accountId)
	defer logger.Exit(ctx, "chatinvites.CreateForChannel", &invite, &err)

	canInvite, err := chats.HasChannelPermission(ctx, channelId, accountId, chatspg.PermInvite)
	if err != nil {
		return invite, "", err
	}
	if !canInvite {
		return invite, "", chats.ErrNotChannelAdmin
	}

	now := time.Now()
	for attempt := 0; ; attempt++ {
		var pinHash string
		plaintextPin, pinHash = generatePinHash()
		invite = chatinvitespg.ChatInvite{
			Id:               uuid.NewString(),
			PinHash:          pinHash,
			ChannelId:        channelId,
			InviterAccountId: accountId,
			ExpiresAt:        now.Add(expiryDuration),
			CreateTime:       now,
		}

		err = gormdb.Create(ctx, &invite)
		if err == nil {
			break
		}
		if !pinhash.IsUniqueViolation(err) || attempt >= maxPinGenerationAttempts-1 {
			return chatinvitespg.ChatInvite{}, "", err
		}
	}

	return invite, plaintextPin, nil
}
