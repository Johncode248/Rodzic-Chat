package chats

import (
	"context"
	"errors"
	"strings"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/google/uuid"
)

var (
	ErrNotAMember        = errors.New("that person is not a member of this chat")
	ErrRoleNotAssignable = errors.New("role must be member or moderator")
	ErrBanTooLong        = errors.New("a temporary ban can last at most 366 days -- use 0 for a ban until lifted")
)

// BannedError is returned to a banned member who tries to use the chat.
type BannedError struct{ Until *time.Time }

func (e *BannedError) Error() string {
	if e.Until == nil {
		return "Masz blokadę w tym czacie – administrator musi ją zdjąć."
	}
	return "Masz blokadę w tym czacie do " + e.Until.Local().Format("02.01.2006 15:04") + "."
}

func banQuery(channelId, accountId string) databasepg.Query {
	return databasepg.Query{Groups: []databasepg.Group{{Filters: []databasepg.Filter{
		{Key: "channel_id", Operator: "=", Value: channelId},
		{Key: "account_id", Operator: "=", Value: accountId},
	}}}}
}

// ActiveBan returns the ban currently in force for accountId in channelId, or nil.
func ActiveBan(ctx context.Context, channelId, accountId string) (*chatspg.ChannelBan, error) {
	bans, err := gormdb.Select[chatspg.ChannelBan](ctx, banQuery(channelId, accountId))
	if err != nil {
		return nil, err
	}
	now := time.Now()
	for i := range bans {
		if bans[i].Active(now) {
			return &bans[i], nil
		}
	}
	return nil, nil
}

// CheckNotBanned returns a *BannedError when accountId is banned from channelId (else nil).
func CheckNotBanned(ctx context.Context, channelId, accountId string) error {
	ban, err := ActiveBan(ctx, channelId, accountId)
	if err != nil {
		return err
	}
	if ban != nil {
		return &BannedError{Until: ban.Until}
	}
	return nil
}

// ListActiveBans returns the bans currently in force in channelId.
func ListActiveBans(ctx context.Context, channelId string) ([]chatspg.ChannelBan, error) {
	bans, err := gormdb.Select[chatspg.ChannelBan](ctx, databasepg.Query{Groups: []databasepg.Group{{Filters: []databasepg.Filter{
		{Key: "channel_id", Operator: "=", Value: channelId},
	}}}})
	if err != nil {
		return nil, err
	}
	now := time.Now()
	active := bans[:0]
	for _, b := range bans {
		if b.Active(now) {
			active = append(active, b)
		}
	}
	return active, nil
}

// MemberPermissions returns callerAccountId's effective permissions in channelId (0 for a non-member).
func MemberPermissions(ctx context.Context, channelId, accountId string) (chatspg.ChannelPermission, error) {
	m, err := getMember(ctx, channelId, accountId)
	if err != nil || m == nil {
		return 0, err
	}
	return m.EffectivePermissions(), nil
}

// HasChannelPermission reports whether accountId has `bit` in channelId.
func HasChannelPermission(ctx context.Context, channelId, accountId string, bit chatspg.ChannelPermission) (bool, error) {
	p, err := MemberPermissions(ctx, channelId, accountId)
	return p.Has(bit), err
}

// canModerate: nobody touches an admin; the admin touches everyone else; a moderator (or anyone
// holding just a permission) only acts on PLAIN members -- never on another moderator.
func canModerate(caller, target *chatspg.ChannelMember) bool {
	if target.Role == chatspg.ChannelMemberRoleAdmin {
		return false
	}
	if caller.Role == chatspg.ChannelMemberRoleAdmin {
		return true
	}
	return target.Role == chatspg.ChannelMemberRoleMember
}

// requireModeration checks that the caller holds `bit` and outranks `target`.
func requireModeration(ctx context.Context, channelId, callerAccountId string, target *chatspg.ChannelMember, bit chatspg.ChannelPermission) error {
	caller, err := getMember(ctx, channelId, callerAccountId)
	if err != nil {
		return err
	}
	if caller == nil || !caller.EffectivePermissions().Has(bit) {
		return ErrNoPermission
	}
	if !canModerate(caller, target) {
		if target.Role == chatspg.ChannelMemberRoleAdmin {
			return ErrCannotRemoveAdmin
		}
		return ErrCannotModerateTarget
	}
	return nil
}

// SetMemberRole changes another member's role (member/moderator) and/or individually granted
// permissions. Admin only; the admin's own entry can't be edited.
func SetMemberRole(ctx context.Context, req chatspg.SetMemberRoleRequest, callerAccountId string) (err error) {
	ctx = logger.Enter(ctx, "chats.SetMemberRole", req, callerAccountId)
	defer logger.Exit(ctx, "chats.SetMemberRole", &err)

	isAdmin, err := IsChannelAdmin(ctx, req.ChannelId, callerAccountId)
	if err != nil {
		return err
	}
	if !isAdmin {
		return ErrNotChannelAdmin
	}
	target, err := getMember(ctx, req.ChannelId, req.AccountId)
	if err != nil {
		return err
	}
	if target == nil {
		return ErrNotAMember
	}
	if target.Role == chatspg.ChannelMemberRoleAdmin {
		return ErrCannotModerateTarget
	}

	if req.Role != nil {
		if *req.Role != chatspg.ChannelMemberRoleMember && *req.Role != chatspg.ChannelMemberRoleModerator {
			return ErrRoleNotAssignable
		}
		target.Role = *req.Role
	}
	if req.Permissions != nil {
		// Unknown bits are dropped, never stored.
		target.Permissions = *req.Permissions & chatspg.AllChannelPermissions
	}
	return gormdb.Update(ctx, target, databasepg.Query{Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: target.Id}}}}})
}

// Ban blocks a member for req.DurationMinutes (0 = until lifted). Replaces any previous ban.
func Ban(ctx context.Context, req chatspg.BanRequest, callerAccountId string) (err error) {
	ctx = logger.Enter(ctx, "chats.Ban", req, callerAccountId)
	defer logger.Exit(ctx, "chats.Ban", &err)

	if callerAccountId == req.AccountId {
		return ErrCannotModerateTarget
	}
	if req.DurationMinutes < 0 || req.DurationMinutes > 366*24*60 {
		return ErrBanTooLong
	}
	target, err := getMember(ctx, req.ChannelId, req.AccountId)
	if err != nil {
		return err
	}
	if target == nil {
		return ErrNotAMember
	}
	if err = requireModeration(ctx, req.ChannelId, callerAccountId, target, chatspg.PermBan); err != nil {
		return err
	}

	if _, err = gormdb.Delete[chatspg.ChannelBan](ctx, banQuery(req.ChannelId, req.AccountId)); err != nil {
		return err
	}
	now := time.Now()
	ban := chatspg.ChannelBan{
		Id: uuid.NewString(), ChannelId: req.ChannelId, AccountId: req.AccountId, BannedBy: callerAccountId,
		Reason: strings.TrimSpace(req.Reason), CreateTime: now,
	}
	if len([]rune(ban.Reason)) > 200 {
		ban.Reason = string([]rune(ban.Reason)[:200])
	}
	if req.DurationMinutes > 0 {
		until := now.Add(time.Duration(req.DurationMinutes) * time.Minute)
		ban.Until = &until
	}
	return gormdb.Create(ctx, &ban)
}

// Unban lifts a member's ban (PermBan).
func Unban(ctx context.Context, channelId, targetAccountId, callerAccountId string) (err error) {
	ctx = logger.Enter(ctx, "chats.Unban", channelId, targetAccountId, callerAccountId)
	defer logger.Exit(ctx, "chats.Unban", &err)

	ok, err := HasChannelPermission(ctx, channelId, callerAccountId, chatspg.PermBan)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNoPermission
	}
	_, err = gormdb.Delete[chatspg.ChannelBan](ctx, banQuery(channelId, targetAccountId))
	return err
}
