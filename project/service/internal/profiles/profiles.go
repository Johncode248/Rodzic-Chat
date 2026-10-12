package profiles

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	authorizationpg "github.com/Johncode248/Rodzic-Chat/project/common/authorization"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
	profilespg "github.com/Johncode248/Rodzic-Chat/project/common/profiles"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/authorization"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/credentials"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/organizations"
	"github.com/google/uuid"
)

var (
	ErrKindExists   = errors.New("you already have a profile of this kind")
	ErrBadKind      = errors.New("unknown profile kind")
	ErrNotYourGroup = errors.New("that profile does not belong to your login")
	ErrMissing      = errors.New("name, phone and organization name are required")
)

func byId(key, value string) databasepg.Query {
	return databasepg.Query{Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: key, Operator: "=", Value: value}}}}}
}

func load(ctx context.Context, id string) (accountspg.Account, error) {
	list, err := gormdb.Select[accountspg.Account](ctx, byId("id", id))
	if err != nil {
		return accountspg.Account{}, err
	}
	if len(list) == 0 {
		return accountspg.Account{}, errors.New("account not found")
	}
	return list[0], nil
}

// group returns every account of the caller's login (the caller alone when it has no group yet).
func group(ctx context.Context, caller accountspg.Account) ([]accountspg.Account, error) {
	if caller.LoginGroupId == "" {
		return []accountspg.Account{caller}, nil
	}
	return gormdb.Select[accountspg.Account](ctx, byId("login_group_id", caller.LoginGroupId))
}

func kindOf(ctx context.Context, a accountspg.Account) (kind, orgName string) {
	if a.AccountType == accountspg.AccountTypeParent {
		return profilespg.KindParent, ""
	}
	kind = profilespg.KindDirector
	if orgs, err := gormdb.Select[organizationspg.Organization](ctx, byId("id", a.OrganizationId)); err == nil && len(orgs) > 0 {
		orgName = orgs[0].Name
		if orgs[0].Type == organizationspg.OrganizationTypeExtracurricular {
			kind = profilespg.KindOrganizer
		}
	}
	return kind, orgName
}

func List(ctx context.Context, accountId string) (res profilespg.ListResponse, err error) {
	ctx = logger.Enter(ctx, "profiles.List", accountId)
	defer logger.Exit(ctx, "profiles.List", &res, &err)

	res.Profiles = []profilespg.Profile{}
	caller, err := load(ctx, accountId)
	if err != nil {
		return res, err
	}
	members, err := group(ctx, caller)
	if err != nil {
		return res, err
	}
	for _, m := range members {
		if m.AccountType == accountspg.AccountTypeTeacher || m.AccountType == accountspg.AccountTypeCoDirector {
			continue // staff hired by a director are their own logins, not the user's own profiles
		}
		kind, orgName := kindOf(ctx, m)
		res.Profiles = append(res.Profiles, profilespg.Profile{AccountId: m.Id, Kind: kind, Name: m.Name, OrganizationName: orgName, Current: m.Id == accountId})
	}
	return res, nil
}

// Add creates another profile (a new account) in the caller's login group. Its stored e-mail is a synthetic
// alias -- nobody logs in with it; the user reaches it via Switch from an existing profile.
func Add(ctx context.Context, accountId string, req profilespg.AddRequest) (res profilespg.ListResponse, err error) {
	ctx = logger.Enter(ctx, "profiles.Add", accountId, req)
	defer logger.Exit(ctx, "profiles.Add", &res, &err)

	if req.Kind != profilespg.KindParent && req.Kind != profilespg.KindDirector && req.Kind != profilespg.KindOrganizer {
		return res, ErrBadKind
	}
	caller, err := load(ctx, accountId)
	if err != nil {
		return res, err
	}
	existing, err := List(ctx, accountId)
	if err != nil {
		return res, err
	}
	for _, p := range existing.Profiles {
		if p.Kind == req.Kind {
			return res, ErrKindExists
		}
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = caller.Name
	}
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		phone = caller.Phone
	}

	accountType := accountspg.AccountTypeParent
	organizationId := ""
	if req.Kind != profilespg.KindParent {
		if strings.TrimSpace(req.OrganizationName) == "" || name == "" {
			return res, ErrMissing
		}
		orgType := organizationspg.OrganizationTypeExtracurricular
		if req.Kind == profilespg.KindDirector {
			mapped, mapErr := authorizationpg.MapOrganizationType(req.OrganizationType)
			if mapErr != nil {
				return res, mapErr
			}
			orgType = mapped
		}
		org, orgErr := organizations.Create(ctx, organizationspg.CreateArgument{Name: strings.TrimSpace(req.OrganizationName), Type: orgType, ChildrenLimit: 50})
		if orgErr != nil {
			return res, orgErr
		}
		organizationId = org.Organization.Id
		if err = authorization.SeedOrganizationFeatures(ctx, organizationId); err != nil {
			return res, err
		}
		accountType = accountspg.AccountTypeDirector
	}

	groupId := caller.LoginGroupId
	if groupId == "" {
		groupId = caller.Id
		caller.LoginGroupId = groupId
		caller.UpdateTime = time.Now()
		if err = gormdb.Update(ctx, &caller, byId("id", caller.Id)); err != nil {
			return res, err
		}
	}

	local, domain := caller.Email, "example.invalid"
	if i := strings.LastIndex(caller.Email, "@"); i > 0 {
		local, domain = caller.Email[:i], caller.Email[i+1:]
	}
	alias := fmt.Sprintf("%s+%s.%s@%s", local, req.Kind, uuid.NewString()[:6], domain)
	created, err := accounts.Create(ctx, accountspg.CreateArgument{
		OrganizationId: organizationId, Name: name, Email: alias, Phone: phone, AccountType: accountType, Status: accountspg.StatusEnabled,
	})
	if err != nil {
		return res, err
	}
	fresh := created.Account
	fresh.LoginGroupId = groupId
	fresh.UpdateTime = time.Now()
	if err = gormdb.Update(ctx, &fresh, byId("id", fresh.Id)); err != nil {
		return res, err
	}
	return List(ctx, accountId)
}

// Switch mints a fresh session for another profile of the caller's login; the client stores it next to the old one.
func Switch(ctx context.Context, accountId, targetId string) (res profilespg.TokensResponse, err error) {
	ctx = logger.Enter(ctx, "profiles.Switch", accountId, targetId)
	defer logger.Exit(ctx, "profiles.Switch", &res, &err)

	caller, err := load(ctx, accountId)
	if err != nil {
		return res, err
	}
	target, err := load(ctx, targetId)
	if err != nil {
		return res, err
	}
	if caller.LoginGroupId == "" || caller.LoginGroupId != target.LoginGroupId {
		return res, ErrNotYourGroup
	}
	issued, err := credentials.IssueTokens(ctx, credentials.IssueTokensArgument{AccountId: target.Id, Email: target.Email})
	if err != nil {
		return res, err
	}
	return profilespg.TokensResponse{AccessToken: issued.AccessToken, RefreshToken: issued.RefreshToken, ExpiresIn: int(credentials.AccessTokenTTL.Seconds())}, nil
}
