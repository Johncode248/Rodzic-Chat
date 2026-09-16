package teacherinvites

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/expressions"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	credentialspg "github.com/Johncode248/Rodzic-Chat/project/common/credentials"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
	teacherinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/teacherinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/credentials"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/groups"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/organizations"
	"github.com/google/uuid"

	groupspg "github.com/Johncode248/Rodzic-Chat/project/common/groups"
)

var ErrInviteNotFound = errors.New("teacher invite not found")
var ErrInviteExpired = errors.New("this teacher invite has expired")
var ErrInviteAlreadyUsed = errors.New("this teacher invite has already been used")
var ErrEmailAlreadyUsed = errors.New("an account with this email already exists")

func byToken(token string) databasepg.Query {
	return databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "token", Operator: "=", Value: token}}},
		},
	}
}

// GetPublicInfo resolves a token for the public teacher-invite page: the
// org/group names to display. Does NOT error on an expired/used/missing
// invite -- Usable=false + a Reason renders more nicely on a static page
// than a raw HTTP error.
func GetPublicInfo(ctx context.Context, token string) (res teacherinvitespg.PublicInfoResponse, err error) {
	ctx = logger.Enter(ctx, "teacherinvites.GetPublicInfo", token)
	defer logger.Exit(ctx, "teacherinvites.GetPublicInfo", &res, &err)

	inviteResult, err := Select(ctx, teacherinvitespg.SelectArgument{Query: byToken(token)})
	if err != nil {
		return res, err
	}
	if len(inviteResult.TeacherInvites) == 0 {
		return teacherinvitespg.PublicInfoResponse{Usable: false, Reason: "not_found"}, nil
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

	return teacherinvitespg.PublicInfoResponse{
		OrganizationName: orgName,
		GroupName:        groupName,
		Usable:           usable,
		Reason:           reason,
	}, nil
}

// PublicRegister completes a teacher's account through a director-generated
// invite link: the invite must still be usable (not expired, not already
// used), the email must not already belong to any account (unlike the
// parent flow, a teacher account is always single-org -- no auto-attach),
// and once created the invite is marked used so it can't be replayed.
func PublicRegister(ctx context.Context, argument teacherinvitespg.PublicRegisterArgument) (res teacherinvitespg.PublicRegisterResponse, err error) {
	ctx = logger.Enter(ctx, "teacherinvites.PublicRegister", argument)
	defer logger.Exit(ctx, "teacherinvites.PublicRegister", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	inviteResult, err := Select(ctx, teacherinvitespg.SelectArgument{Query: byToken(argument.Token)})
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

	emailTaken, err := accounts.Select(ctx, accountspg.SelectArgument{
		Query: databasepg.Query{
			Limit:  1,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "login", Operator: "=", Value: argument.Email}}}},
		},
	})
	if err != nil {
		return res, err
	}
	if len(emailTaken.Accounts) > 0 {
		return res, ErrEmailAlreadyUsed
	}

	accountType, err := authorizationpg.MapAccountType(expressions.Select(invite.AccountType != "", invite.AccountType, "teacher"))
	if err != nil {
		return res, err
	}

	accountId := uuid.NewString()

	if err = routines.Select(
		func() error {
			_, err := accounts.Create(ctx, accountspg.CreateArgument{
				Id:             accountId,
				OrganizationId: invite.OrganizationId,
				Name:           argument.Name,
				Login:          argument.Email,
				Email:          argument.Email,
				AccountType:    accountType,
				Status:         accountspg.StatusEnabled,
			})
			return err
		},
		func() error {
			_, err := credentials.Create(ctx, credentialspg.CreateArgument{
				UserId:   accountId,
				Password: argument.Password,
			})
			return err
		},
	); err != nil {
		// One of the two may have already committed (they run
		// concurrently) -- clean up both rather than leaving an orphaned
		// Account with no Credential (permanently unusable, email stays
		// "taken" forever) or vice versa.
		_ = rollbackPartialAccountCreate(ctx, accountId)
		return res, err
	}

	if invite.GroupId != "" && accountType == accountspg.AccountTypeTeacher {
		if err = AssignTeacherToGroup(ctx, invite.GroupId, accountId, invite.OrganizationId); err != nil {
			return res, err
		}
	}

	// Atomic conditional update, not a blind write: two concurrent
	// requests for the same single-use token could otherwise both pass
	// the "UsedAt == nil" check above and both create an account before
	// either commits this update. Gating the UPDATE itself on "still
	// unused" and checking RowsAffected makes this the real enforcement
	// point -- only the first caller to actually reach it wins.
	now := time.Now()
	updateResult := gormdb.GormDB.WithContext(ctx).
		Model(&teacherinvitespg.TeacherInvite{}).
		Where("token = ? AND used_at IS NULL", argument.Token).
		Update("used_at", now)
	if updateResult.Error != nil {
		return res, updateResult.Error
	}
	if updateResult.RowsAffected == 0 {
		_ = rollbackPartialAccountCreate(ctx, accountId)
		return res, ErrInviteAlreadyUsed
	}

	return teacherinvitespg.PublicRegisterResponse{
		Message: "Konto utworzone. Możesz się teraz zalogować w aplikacji mobilnej Rodzic Chat, używając podanego adresu e-mail i hasła.",
	}, nil
}

// rollbackPartialAccountCreate cleans up after a failed concurrent
// Account+Credential creation above -- one of the two may have committed
// before the other failed. Deleting a row that was never created is a
// no-op given the query-based delete, so it's safe to call both
// unconditionally.
func rollbackPartialAccountCreate(ctx context.Context, accountId string) error {
	return routines.Select(
		func() error {
			return accounts.FinalizeDeleteMutation(ctx, &accountspg.DeleteMutation{
				Foundation: accountspg.DeleteArgument{
					Query: databasepg.Query{
						Groups: []databasepg.Group{
							{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: accountId}}},
						},
					},
				},
			})
		},
		func() error {
			return credentials.ProcessDeleteMutation(ctx, &credentialspg.DeleteMutation{
				Foundation: credentialspg.DeleteArgument{
					Query: databasepg.Query{
						Groups: []databasepg.Group{
							{Filters: []databasepg.Filter{{Key: "user_id", Operator: "=", Value: accountId}}},
						},
					},
				},
			})
		},
	)
}
