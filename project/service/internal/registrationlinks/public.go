package registrationlinks

import (
	"context"
	"errors"
	"strings"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountorganizationspg "github.com/Johncode248/Rodzic-Chat/project/common/accountorganizations"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	credentialspg "github.com/Johncode248/Rodzic-Chat/project/common/credentials"
	groupspg "github.com/Johncode248/Rodzic-Chat/project/common/groups"
	guardianclaimspg "github.com/Johncode248/Rodzic-Chat/project/common/guardianclaims"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
	registrationlinkspg "github.com/Johncode248/Rodzic-Chat/project/common/registrationlinks"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accountorganizations"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/credentials"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/groups"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/guardianclaims"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/organizations"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentchild"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrLinkNotFound = errors.New("registration link not found")
var ErrLinkExpired = errors.New("this registration link has expired")
var ErrLinkExhausted = errors.New("this registration link has reached its registration limit")
var ErrChildAlreadyHasTwoParents = errors.New("this child already has two linked parent accounts")
var ErrNameAlreadyRegisteredInGroup = errors.New("a parent with this name is already registered for this group")
var ErrWrongPassword = errors.New("an account with this email already exists -- the password does not match it (use password recovery if you forgot it)")
var ErrChildBirthDateRequired = errors.New("this group requires the child's birth date to register")

func byToken(token string) databasepg.Query {
	return databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "token", Operator: "=", Value: token}}},
		},
	}
}

// GetPublicInfo resolves a token for the public registration page: the
// org/group names to display -- the parent types their child's name
// themselves (see PublicRegister/findOrCreateChild), there's no picker. It
// intentionally does NOT error on an expired/exhausted/missing link —
// Usable=false + a Reason is easier for a static page to render nicely
// than a raw HTTP error.
func GetPublicInfo(ctx context.Context, token string) (res registrationlinkspg.PublicInfoResponse, err error) {
	ctx = logger.Enter(ctx, "registrationlinks.GetPublicInfo", token)
	defer logger.Exit(ctx, "registrationlinks.GetPublicInfo", &res, &err)

	linkResult, err := Select(ctx, registrationlinkspg.SelectArgument{Query: byToken(token)})
	if err != nil {
		return res, err
	}
	if len(linkResult.RegistrationLinks) == 0 {
		return registrationlinkspg.PublicInfoResponse{Usable: false, Reason: "not_found"}, nil
	}
	link := linkResult.RegistrationLinks[0]

	orgResult, err := organizations.Select(ctx, organizationspg.SelectArgument{
		Query: databasepg.Query{
			Limit:  1,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: link.OrganizationId}}}},
		},
	})
	if err != nil {
		return res, err
	}

	groupResult, err := groups.Select(ctx, groupspg.SelectArgument{
		Query: databasepg.Query{
			Limit:  1,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: link.GroupId}}}},
		},
	})
	if err != nil {
		return res, err
	}

	usable := link.IsUsable(time.Now())
	reason := ""
	if !usable {
		if time.Now().After(link.ExpiresAt) {
			reason = "expired"
		} else {
			reason = "exhausted"
		}
	}

	orgName, groupName := "", ""
	if len(orgResult.Organizations) > 0 {
		orgName = orgResult.Organizations[0].Name
	}
	if len(groupResult.Groups) > 0 {
		groupName = groupResult.Groups[0].Name
	}

	return registrationlinkspg.PublicInfoResponse{
		OrganizationName: orgName,
		GroupName:        groupName,
		Usable:           usable,
		Reason:           reason,
	}, nil
}

// PublicRegister completes a parent's self-registration through a link.
// Every business constraint from the product spec is enforced here: the
// link must still be usable (not expired, under its registration cap), the
// named child can't already have two linked parents, and the full name
// must not already be registered for this group (no two "Magda Kowalska"
// under the same group). The parent names the child themselves — staff
// never pre-create children for this flow, see resolveChild.
//
// Two things a registering parent can never prove just by filling out this
// public, unauthenticated form are handled explicitly rather than trusted:
//  1. Owning the email they typed, if it already belongs to an account
//     (see the joiningExistingAccount branch — the existing account's
//     password is verified, not skipped).
//  2. Actually being a parent of an existing child they matched by name
//     (see resolveChild / guardianclaims — a birth-date mismatch, when the
//     group requires verification, defers to director approval instead of
//     auto-linking).
func PublicRegister(ctx context.Context, argument registrationlinkspg.PublicRegisterArgument) (res registrationlinkspg.PublicRegisterResponse, err error) {
	ctx = logger.Enter(ctx, "registrationlinks.PublicRegister", argument)
	defer logger.Exit(ctx, "registrationlinks.PublicRegister", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	linkResult, err := Select(ctx, registrationlinkspg.SelectArgument{Query: byToken(argument.Token)})
	if err != nil {
		return res, err
	}
	if len(linkResult.RegistrationLinks) == 0 {
		return res, ErrLinkNotFound
	}
	link := linkResult.RegistrationLinks[0]

	if time.Now().After(link.ExpiresAt) {
		return res, ErrLinkExpired
	}
	if link.UsedCount >= link.MaxRegistrations {
		return res, ErrLinkExhausted
	}

	groupResult, err := groups.Select(ctx, groupspg.SelectArgument{
		Query: databasepg.Query{
			Limit:  1,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: link.GroupId}}}},
		},
	})
	if err != nil {
		return res, err
	}
	requireBirthDate := len(groupResult.Groups) > 0 && groupResult.Groups[0].RequireChildBirthDateVerification

	if requireBirthDate && argument.ChildBirthDate == nil {
		return res, ErrChildBirthDateRequired
	}

	resolved, err := resolveChild(ctx, link.OrganizationId, link.GroupId, argument.ChildFirstName, argument.ChildLastName, argument.ChildBirthDate)
	if err != nil {
		return res, err
	}

	existingLinksForChild, err := parentchild.Select(ctx, parentchildpg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "child_id", Operator: "=", Value: resolved.childId}}}},
		},
	})
	if err != nil {
		return res, err
	}
	if len(existingLinksForChild.ParentChildren) >= 2 {
		return res, ErrChildAlreadyHasTwoParents
	}

	effectiveEmail := argument.EffectiveEmail()
	viaProvider := argument.VerifiedProviderEmail != ""

	emailTaken, err := accounts.Select(ctx, accountspg.SelectArgument{
		Query: databasepg.Query{
			Limit:  1,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "login", Operator: "=", Value: effectiveEmail}}}},
		},
	})
	if err != nil {
		return res, err
	}
	// An existing account joining a second (or further) facility: per
	// product decision, this auto-attaches to the existing account by
	// matching email rather than erroring or creating a fragmented second
	// account -- see accountorganizationspg.AccountOrganization. Ownership
	// of that email is still proven, not skipped -- either the submitted
	// password verified against that account's real credential, or (via
	// provider sign-in) the fact that Google/Apple themselves already
	// vouched for this exact email before this function ever ran (see
	// oauthverify) -- an even stronger guarantee than a password, so it's
	// accepted the same way. Either way this can't be used to attach to an
	// account whose email you merely know, not own -- see ErrWrongPassword.
	var accountId string
	joiningExistingAccount := len(emailTaken.Accounts) > 0

	if joiningExistingAccount {
		accountId = emailTaken.Accounts[0].Id

		if !viaProvider {
			credentialResult, err := credentials.Select(ctx, credentialspg.SelectArgument{
				Query: databasepg.Query{
					Limit:  1,
					Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "user_id", Operator: "=", Value: accountId}}}},
				},
			})
			if err != nil {
				return res, err
			}
			if len(credentialResult.Credentials) == 0 ||
				bcrypt.CompareHashAndPassword([]byte(credentialResult.Credentials[0].Password), []byte(argument.Password)) != nil {
				return res, ErrWrongPassword
			}
		}
	} else {
		nameTaken, err := nameAlreadyRegisteredInGroup(ctx, link.GroupId, argument.Name)
		if err != nil {
			return res, err
		}
		if nameTaken {
			return res, ErrNameAlreadyRegisteredInGroup
		}

		accountId = uuid.NewString()

		createAccount := func() error {
			_, err := accounts.Create(ctx, accountspg.CreateArgument{
				Id:             accountId,
				OrganizationId: link.OrganizationId,
				Name:           argument.Name,
				Login:          effectiveEmail,
				Email:          effectiveEmail,
				AccountType:    accountspg.AccountTypeParent,
				Status:         accountspg.StatusEnabled,
			})
			return err
		}

		if viaProvider {
			// No Credential row at all: this account can only ever log in
			// via the same verified provider email (or by setting a
			// password later through /public/password-recovery, which
			// works even without an existing Credential).
			if err = createAccount(); err != nil {
				_ = rollbackPartialAccountCreate(ctx, accountId)
				return res, err
			}
		} else if err = routines.Select(
			createAccount,
			func() error {
				_, err := credentials.Create(ctx, credentialspg.CreateArgument{
					UserId:   accountId,
					Password: argument.Password,
				})
				return err
			},
		); err != nil {
			// One of the two may have already committed (they run
			// concurrently) -- clean up both unconditionally rather than
			// leaving an orphaned Account with no Credential (permanently
			// unusable, and its email/login stays "taken" forever) or vice
			// versa. Best-effort: deleting a row that was never created is
			// a no-op given the query-based delete.
			_ = rollbackPartialAccountCreate(ctx, accountId)
			return res, err
		}
	}

	if err = ensureOrganizationMembership(ctx, accountId, link.OrganizationId, !joiningExistingAccount); err != nil {
		return res, err
	}

	if resolved.needsDirectorApproval {
		if _, err = guardianclaims.Create(ctx, guardianclaimspg.CreateArgument{
			OrganizationId:       link.OrganizationId,
			ChildId:              resolved.childId,
			ClaimantId:           accountId,
			ClaimantName:         argument.Name,
			ClaimantEmail:        effectiveEmail,
			ClaimantBirthDate:    *argument.ChildBirthDate,
			ChildBirthDateOnFile: resolved.childBirthDateOnFile,
		}); err != nil {
			return res, err
		}
	} else {
		// The count-then-insert two-parent check above is only an early,
		// informational rejection -- two concurrent registrations naming
		// the same child can both pass it before either commits. Re-check
		// atomically inside a transaction serialized per-child via
		// pg_advisory_xact_lock (same pattern as chatinvites.Create's
		// per-account channel-limit check) immediately before the actual
		// insert, so the real enforcement point can't be raced.
		err = gormdb.GormDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", resolved.childId).Error; err != nil {
				return err
			}

			var parentCount int64
			if err := tx.Model(&parentchildpg.ParentChild{}).Where("child_id = ?", resolved.childId).Count(&parentCount).Error; err != nil {
				return err
			}
			if parentCount >= 2 {
				return ErrChildAlreadyHasTwoParents
			}

			pc := parentchildpg.CreateArgument{
				ParentId:  accountId,
				ChildId:   resolved.childId,
				Relation:  parentchildpg.RelationTypeGuardian,
				CanPickup: true,
			}.Populate()
			return tx.Create(&pc).Error
		})
		if err != nil {
			return res, err
		}
	}

	// Atomic conditional increment, not a blind read-then-write: two
	// concurrent registrations at the last remaining slot could otherwise
	// both read the same pre-increment UsedCount, both pass the earlier
	// informational check, and both commit, exceeding MaxRegistrations.
	// Gating the UPDATE itself on "still under the cap" and checking
	// RowsAffected makes this the real enforcement point -- only as many
	// concurrent callers as there are remaining slots can ever succeed.
	updateResult := gormdb.GormDB.WithContext(ctx).
		Model(&registrationlinkspg.RegistrationLink{}).
		Where("token = ? AND used_count < max_registrations", argument.Token).
		Update("used_count", gorm.Expr("used_count + 1"))
	if updateResult.Error != nil {
		return res, updateResult.Error
	}
	if updateResult.RowsAffected == 0 {
		return res, ErrLinkExhausted
	}

	if resolved.needsDirectorApproval {
		return registrationlinkspg.PublicRegisterResponse{
			Message:         "Konto utworzone. Ponieważ podana data urodzenia dziecka nie zgadza się z tym, co widnieje w systemie, dostęp do dziecka wymaga zatwierdzenia przez dyrekcję placówki. Możesz się już zalogować, ale dziecko pojawi się dopiero po zatwierdzeniu.",
			PendingApproval: true,
		}, nil
	}

	return registrationlinkspg.PublicRegisterResponse{
		Message: "Konto utworzone. Możesz się teraz zalogować w aplikacji mobilnej Rodzic Chat, używając podanego adresu e-mail i hasła.",
	}, nil
}

// resolvedChild is resolveChild's result: which Child row to use, and
// whether linking to it must go through director approval instead of
// happening immediately.
type resolvedChild struct {
	childId               string
	needsDirectorApproval bool
	childBirthDateOnFile  time.Time
}

// resolveChild resolves the child a registering parent named to an actual
// Child row: a case-insensitive, trimmed exact match against an existing
// child in the same group is reused (this is how a second parent -- e.g.
// the other guardian -- registering separately via the same link attaches
// to the SAME child instead of creating a duplicate), otherwise a new
// Child is created. Staff never pre-create children for this flow — the
// parent is the one naming the child, per product decision.
//
// When matching an EXISTING child and the group requires birth-date
// verification, the caller-provided date must agree with what's on file
// (i.e. what the first parent provided) for this to auto-link; a mismatch
// sets needsDirectorApproval instead of erroring outright, since the
// mismatch could be an honest typo, not necessarily an impostor -- the
// director reviews the actual submitted values and decides (see
// project/common/guardianclaims).
func resolveChild(ctx context.Context, organizationId, groupId, firstName, lastName string, providedBirthDate *time.Time) (resolvedChild, error) {
	childrenResult, err := children.Select(ctx, childrenpg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "group_id", Operator: "=", Value: groupId}}}},
		},
	})
	if err != nil {
		return resolvedChild{}, err
	}

	normalizedFirst := strings.ToLower(strings.TrimSpace(firstName))
	normalizedLast := strings.ToLower(strings.TrimSpace(lastName))
	for _, child := range childrenResult.Children {
		if strings.ToLower(strings.TrimSpace(child.FirstName)) != normalizedFirst ||
			strings.ToLower(strings.TrimSpace(child.LastName)) != normalizedLast {
			continue
		}

		if providedBirthDate == nil {
			// Group doesn't require verification -- today's behavior,
			// unchanged: match by name alone.
			return resolvedChild{childId: child.Id}, nil
		}

		if providedBirthDate.Format("2006-01-02") == child.BirthDate.Format("2006-01-02") {
			return resolvedChild{childId: child.Id}, nil
		}

		return resolvedChild{
			childId:               child.Id,
			needsDirectorApproval: true,
			childBirthDateOnFile:  child.BirthDate,
		}, nil
	}

	createArgument := childrenpg.CreateArgument{
		OrganizationId: organizationId,
		GroupId:        groupId,
		FirstName:      firstName,
		LastName:       lastName,
	}
	if providedBirthDate != nil {
		createArgument.BirthDate = *providedBirthDate
	}

	createResult, err := children.Create(ctx, createArgument)
	if err != nil {
		return resolvedChild{}, err
	}

	return resolvedChild{childId: createResult.Child.Id}, nil
}

// nameAlreadyRegisteredInGroup checks every parent already linked to a
// child in groupId for a case-insensitive, whitespace-trimmed match against
// name — the "no two Magda Kowalska in the same group" rule.
func nameAlreadyRegisteredInGroup(ctx context.Context, groupId, name string) (bool, error) {
	childrenResult, err := children.Select(ctx, childrenpg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "group_id", Operator: "=", Value: groupId}}}},
		},
	})
	if err != nil {
		return false, err
	}
	if len(childrenResult.Children) == 0 {
		return false, nil
	}

	childIds := make([]string, 0, len(childrenResult.Children))
	for _, child := range childrenResult.Children {
		childIds = append(childIds, child.Id)
	}

	linksResult, err := parentchild.Select(ctx, parentchildpg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "child_id", Operator: "IN", Value: childIds}}}},
		},
	})
	if err != nil {
		return false, err
	}
	if len(linksResult.ParentChildren) == 0 {
		return false, nil
	}

	parentIds := make([]string, 0, len(linksResult.ParentChildren))
	for _, link := range linksResult.ParentChildren {
		parentIds = append(parentIds, link.ParentId)
	}

	parentsResult, err := accounts.Select(ctx, accountspg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "IN", Value: parentIds}}}},
		},
	})
	if err != nil {
		return false, err
	}

	normalized := strings.ToLower(strings.TrimSpace(name))
	for _, parent := range parentsResult.Accounts {
		if strings.ToLower(strings.TrimSpace(parent.Name)) == normalized {
			return true, nil
		}
	}

	return false, nil
}

// ensureOrganizationMembership creates the AccountOrganization row for a
// parent joining organizationId through a registration link, if one
// doesn't already exist -- covers both a brand-new account's first facility
// (isPrimary=true) and an existing account joining a further facility
// (isPrimary=false), and is a no-op if the parent is already a member (e.g.
// a second child at the same facility via a different group's link).
func ensureOrganizationMembership(ctx context.Context, accountId, organizationId string, isPrimary bool) error {
	existing, err := accountorganizations.Select(ctx, accountorganizationspg.SelectArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{
				{Key: "account_id", Operator: "=", Value: accountId},
				{Key: "organization_id", Operator: "=", Value: organizationId},
			}}},
		},
	})
	if err != nil {
		return err
	}
	if len(existing.AccountOrganizations) > 0 {
		return nil
	}

	_, err = accountorganizations.Create(ctx, accountorganizationspg.CreateArgument{
		AccountId:      accountId,
		OrganizationId: organizationId,
		IsPrimary:      isPrimary,
	})
	return err
}

// rollbackPartialAccountCreate cleans up after a failed concurrent
// Account+Credential creation (see PublicRegister) -- one of the two may
// have committed before the other failed. Deleting a row that was never
// created is a no-op given the query-based delete, so it's safe to call
// both unconditionally.
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
