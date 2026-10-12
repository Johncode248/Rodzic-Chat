package teacherinvites

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/expressions"
	"github.com/Johncode248/Rodzic-Chat/library/pinhash"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
	teacherinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/teacherinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/groups"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/organizations"

	groupspg "github.com/Johncode248/Rodzic-Chat/project/common/groups"
)

var ErrInviteNotFound = errors.New("teacher invite not found")
var ErrInviteExpired = errors.New("this teacher invite has expired")
var ErrInviteAlreadyUsed = errors.New("this teacher invite has already been used")

func byPinHash(pinHash string) databasepg.Query {
	return databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "pin_hash", Operator: "=", Value: pinHash}}},
		},
	}
}

// Lookup resolves a PIN for an already-authenticated caller: the org/group
// names + description to display, before they commit to Confirm. Doesn't
// error on an expired/used/not-found PIN -- Usable=false + a Reason
// renders more nicely on a client than a raw HTTP error.
func Lookup(ctx context.Context, pin string) (res teacherinvitespg.LookupResponse, err error) {
	ctx = logger.Enter(ctx, "teacherinvites.Lookup", "<redacted>")
	defer logger.Exit(ctx, "teacherinvites.Lookup", &res, &err)

	inviteResult, err := Select(ctx, teacherinvitespg.SelectArgument{Query: byPinHash(pinhash.HashPin(pin))})
	if err != nil {
		return res, err
	}
	if len(inviteResult.TeacherInvites) == 0 {
		return teacherinvitespg.LookupResponse{Usable: false, Reason: "not_found"}, nil
	}
	invite := inviteResult.TeacherInvites[0]

	orgResult, err := organizations.Select(ctx, organizationspg.SelectArgument{
		Query: databasepg.Query{
			Limit:  1,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: invite.OrganizationId}}}},
		},
	})
	if err != nil {
		return res, err
	}

	groupName := ""
	if invite.GroupId != "" {
		groupResult, err := groups.Select(ctx, groupspg.SelectArgument{
			Query: databasepg.Query{
				Limit:  1,
				Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: invite.GroupId}}}},
			},
		})
		if err != nil {
			return res, err
		}
		if len(groupResult.Groups) > 0 {
			groupName = groupResult.Groups[0].Name
		}
	}

	usable := invite.IsUsable(time.Now())
	reason := ""
	if !usable {
		if invite.UsedAt != nil {
			reason = "used"
		} else {
			reason = "expired"
		}
	}

	orgName := ""
	if len(orgResult.Organizations) > 0 {
		orgName = orgResult.Organizations[0].Name
	}

	return teacherinvitespg.LookupResponse{
		OrganizationName:  orgName,
		GroupName:         groupName,
		AccountType:       expressions.Select(invite.AccountType != "", invite.AccountType, "teacher"),
		PublicDescription: invite.PublicDescription,
		Usable:            usable,
		Reason:            reason,
	}, nil
}

// Confirm attaches the CALLING account (already authenticated) as staff
// (teacher or co_director, per the invite) of the invite's organization,
// assigns it to the invite's group (teacher invites only), and marks the
// invite used. Unlike the old public register flow this replaces, there is
// no more account-creation/email-uniqueness branch -- the caller IS the
// account.
func Confirm(ctx context.Context, argument teacherinvitespg.ConfirmArgument) (res teacherinvitespg.ConfirmResponse, err error) {
	ctx = logger.Enter(ctx, "teacherinvites.Confirm", argument.AccountId)
	defer logger.Exit(ctx, "teacherinvites.Confirm", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	inviteResult, err := Select(ctx, teacherinvitespg.SelectArgument{Query: byPinHash(pinhash.HashPin(argument.Pin))})
	if err != nil {
		return res, err
	}
	if len(inviteResult.TeacherInvites) == 0 {
		return res, ErrInviteNotFound
	}
	invite := inviteResult.TeacherInvites[0]

	if invite.UsedAt != nil {
		return res, ErrInviteAlreadyUsed
	}
	if time.Now().After(invite.ExpiresAt) {
		return res, ErrInviteExpired
	}

	accountType, err := authorizationpg.MapAccountType(expressions.Select(invite.AccountType != "", invite.AccountType, "teacher"))
	if err != nil {
		return res, err
	}

	// Atomic conditional update, not a blind write: two concurrent
	// requests for the same single-use PIN could otherwise both pass the
	// "UsedAt == nil" check above before either commits this update.
	// Gating the UPDATE itself on "still unused" and checking
	// RowsAffected makes this the real enforcement point -- only the
	// first caller to actually reach it wins.
	now := time.Now()
	claimResult := gormdb.GormDB.WithContext(ctx).
		Model(&teacherinvitespg.TeacherInvite{}).
		Where("id = ? AND used_at IS NULL", invite.Id).
		Update("used_at", now)
	if claimResult.Error != nil {
		return res, claimResult.Error
	}
	if claimResult.RowsAffected == 0 {
		return res, ErrInviteAlreadyUsed
	}

	// Attach the calling account to this organization as staff --
	// promotes its single primary AccountType/OrganizationId (staff
	// accounts, unlike parents, are always single-org).
	updateResult := gormdb.GormDB.WithContext(ctx).
		Model(&accountspg.Account{}).
		Where("id = ?", argument.AccountId).
		Updates(map[string]interface{}{
			"organization_id": invite.OrganizationId,
			"account_type":    accountType,
			"update_time":     now,
		})
	if updateResult.Error != nil {
		return res, updateResult.Error
	}
	if updateResult.RowsAffected == 0 {
		return res, errors.New("account not found")
	}

	if invite.GroupId != "" && accountType == accountspg.AccountTypeTeacher {
		if err = AssignTeacherToGroup(ctx, invite.GroupId, argument.AccountId, invite.OrganizationId); err != nil {
			return res, err
		}
	}

	return teacherinvitespg.ConfirmResponse{
		Message: "Twoje konto zostało przypisane do placówki.",
	}, nil
}
