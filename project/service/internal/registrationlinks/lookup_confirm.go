package registrationlinks

import (
	"context"
	"errors"
	"strings"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/pinhash"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountorganizationspg "github.com/Johncode248/Rodzic-Chat/project/common/accountorganizations"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	groupspg "github.com/Johncode248/Rodzic-Chat/project/common/groups"
	guardianclaimspg "github.com/Johncode248/Rodzic-Chat/project/common/guardianclaims"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
	registrationlinkspg "github.com/Johncode248/Rodzic-Chat/project/common/registrationlinks"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accountorganizations"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/groups"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/guardianclaims"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/organizations"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentchild"
	"gorm.io/gorm"
)

var ErrLinkNotFound = errors.New("registration invite not found")
var ErrLinkExpired = errors.New("this registration invite has expired")
var ErrLinkExhausted = errors.New("this registration invite has reached its registration limit")
var ErrChildAlreadyHasTwoParents = errors.New("this child already has two linked parent accounts")
var ErrChildBirthDateRequired = errors.New("this group requires the child's birth date to register")
var ErrChildNotInGroup = errors.New("the selected child is not in this group")

func byPinHash(pinHash string) databasepg.Query {
	return databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "pin_hash", Operator: "=", Value: pinHash}}},
		},
	}
}

// Lookup resolves a PIN for an already-authenticated parent, before they
// commit to confirming: the org/group names + description to display.
// Doesn't error on an expired/exhausted/not-found PIN -- Usable=false + a
// Reason is easier for a client to render nicely than a raw HTTP error.
func Lookup(ctx context.Context, pin string) (res registrationlinkspg.LookupResponse, err error) {
	ctx = logger.Enter(ctx, "registrationlinks.Lookup", "<redacted>")
	defer logger.Exit(ctx, "registrationlinks.Lookup", &res, &err)

	linkResult, err := Select(ctx, registrationlinkspg.SelectArgument{Query: byPinHash(pinhash.HashPin(pin))})
	if err != nil {
		return res, err
	}
	if len(linkResult.RegistrationLinks) == 0 {
		return registrationlinkspg.LookupResponse{Usable: false, Reason: "not_found"}, nil
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

	var childOptions []registrationlinkspg.PublicChildOption
	if usable {
		childrenResult, err := children.Select(ctx, childrenpg.SelectArgument{
			Query: databasepg.Query{
				Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "group_id", Operator: "=", Value: link.GroupId}}}},
			},
		})
		if err != nil {
			return res, err
		}
		childOptions = make([]registrationlinkspg.PublicChildOption, 0, len(childrenResult.Children))
		for _, c := range childrenResult.Children {
			childOptions = append(childOptions, registrationlinkspg.PublicChildOption{
				Id: c.Id, FirstName: c.FirstName, LastName: c.LastName,
			})
		}
	}

	return registrationlinkspg.LookupResponse{
		OrganizationName:  orgName,
		GroupName:         groupName,
		PublicDescription: link.PublicDescription,
		Usable:            usable,
		Reason:            reason,
		Children:          childOptions,
	}, nil
}

// Confirm completes an already-authenticated parent's attachment to a
// child through a PIN. Every business constraint from the product spec is
// enforced here: the invite must still be usable (not expired, under its
// registration cap), the named child can't already have two linked
// parents. The parent names the child themselves — staff never pre-create
// children for this flow, see resolveChild.
//
// Unlike the old public/anonymous register flow this replaces, the calling
// account already exists and is already authenticated (see
// ConfirmArgument.AccountId) -- there's no more account-creation/OAuth
// branch, and no more "does this email already belong to someone" check,
// since the caller IS the account.
func Confirm(ctx context.Context, argument registrationlinkspg.ConfirmArgument) (res registrationlinkspg.ConfirmResponse, err error) {
	ctx = logger.Enter(ctx, "registrationlinks.Confirm", argument.AccountId)
	defer logger.Exit(ctx, "registrationlinks.Confirm", &res, &err)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	linkResult, err := Select(ctx, registrationlinkspg.SelectArgument{Query: byPinHash(pinhash.HashPin(argument.Pin))})
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

	resolved, err := resolveChild(ctx, link.OrganizationId, link.GroupId, argument.ChildId, argument.ChildFirstName, argument.ChildLastName, argument.ChildBirthDate)
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

	accountId := argument.AccountId

	if err = ensureOrganizationMembership(ctx, accountId, link.OrganizationId); err != nil {
		return res, err
	}

	accountResult, err := gormdb.Select[accountspg.Account](ctx, databasepg.Query{
		Limit:  1,
		Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: accountId}}}},
	})
	if err != nil {
		return res, err
	}
	claimantName := ""
	claimantEmail := ""
	if len(accountResult) > 0 {
		claimantName = accountResult[0].Name
		claimantEmail = accountResult[0].Email
	}

	if resolved.needsDirectorApproval {
		if _, err = guardianclaims.Create(ctx, guardianclaimspg.CreateArgument{
			OrganizationId:       link.OrganizationId,
			ChildId:              resolved.childId,
			ClaimantId:           accountId,
			ClaimantName:         claimantName,
			ClaimantEmail:        claimantEmail,
			ClaimantBirthDate:    *argument.ChildBirthDate,
			ChildBirthDateOnFile: resolved.childBirthDateOnFile,
		}); err != nil {
			return res, err
		}
	} else {
		// The count-then-insert two-parent check above is only an early,
		// informational rejection -- two concurrent confirmations naming
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
				ParentId: accountId,
				ChildId:  resolved.childId,
				Relation: parentchildpg.RelationTypeGuardian,
			}.Populate()
			return tx.Create(&pc).Error
		})
		if err != nil {
			return res, err
		}
	}

	// Atomic conditional increment, not a blind read-then-write: two
	// concurrent confirmations at the last remaining slot could otherwise
	// both read the same pre-increment UsedCount, both pass the earlier
	// informational check, and both commit, exceeding MaxRegistrations.
	// Gating the UPDATE itself on "still under the cap" and checking
	// RowsAffected makes this the real enforcement point -- only as many
	// concurrent callers as there are remaining slots can ever succeed.
	updateResult := gormdb.GormDB.WithContext(ctx).
		Model(&registrationlinkspg.RegistrationLink{}).
		Where("id = ? AND used_count < max_registrations", link.Id).
		Update("used_count", gorm.Expr("used_count + 1"))
	if updateResult.Error != nil {
		return res, updateResult.Error
	}
	if updateResult.RowsAffected == 0 {
		return res, ErrLinkExhausted
	}

	if resolved.needsDirectorApproval {
		return registrationlinkspg.ConfirmResponse{
			Message:         "Ponieważ podana data urodzenia dziecka nie zgadza się z tym, co widnieje w systemie, dostęp do dziecka wymaga zatwierdzenia przez dyrekcję placówki.",
			PendingApproval: true,
		}, nil
	}

	return registrationlinkspg.ConfirmResponse{
		Message: "Dziecko zostało dodane do Twojego konta.",
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

// resolveChild resolves the child a confirming parent named to an actual
// Child row: a case-insensitive, trimmed exact match against an existing
// child in the same group is reused (this is how a second parent -- e.g.
// the other guardian -- confirming separately via the same invite attaches
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
func resolveChild(ctx context.Context, organizationId, groupId, childId, firstName, lastName string, providedBirthDate *time.Time) (resolvedChild, error) {
	if childId != "" {
		return resolveChildById(ctx, groupId, childId, providedBirthDate)
	}

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

// resolveChildById is resolveChild's path for a parent picking one of
// LookupResponse's listed children instead of typing a name -- the same
// birth-date-verification/needsDirectorApproval rule applies as the
// typed-name match above, since picking from a list proves nothing more
// about actually being that child's guardian than typing the name would.
func resolveChildById(ctx context.Context, groupId, childId string, providedBirthDate *time.Time) (resolvedChild, error) {
	childResult, err := children.Select(ctx, childrenpg.SelectArgument{
		Query: databasepg.Query{
			Limit:  1,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: childId}}}},
		},
	})
	if err != nil {
		return resolvedChild{}, err
	}
	if len(childResult.Children) == 0 || childResult.Children[0].GroupId != groupId {
		return resolvedChild{}, ErrChildNotInGroup
	}
	child := childResult.Children[0]

	if providedBirthDate == nil {
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

// ensureOrganizationMembership creates the AccountOrganization row for a
// parent joining organizationId through a registration invite, if one
// doesn't already exist -- covers both this account's first facility
// (isPrimary=true, when it has no memberships yet) and joining a further
// facility (isPrimary=false), and is a no-op if the parent is already a
// member (e.g. a second child at the same facility via a different
// group's invite).
func ensureOrganizationMembership(ctx context.Context, accountId, organizationId string) error {
	existing, err := accountorganizations.Select(ctx, accountorganizationspg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{
				{Key: "account_id", Operator: "=", Value: accountId},
			}}},
		},
	})
	if err != nil {
		return err
	}

	for _, membership := range existing.AccountOrganizations {
		if membership.OrganizationId == organizationId {
			return nil
		}
	}

	_, err = accountorganizations.Create(ctx, accountorganizationspg.CreateArgument{
		AccountId:      accountId,
		OrganizationId: organizationId,
		IsPrimary:      len(existing.AccountOrganizations) == 0,
	})
	return err
}
