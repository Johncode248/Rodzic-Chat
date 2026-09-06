package credentials

import (
	"context"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	sessionspg "github.com/Johncode248/Rodzic-Chat/project/common/sessions"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/sessions"
)

// IssueTokens creates a new Session row for a device/login and mints the
// access+refresh token pair tied to it. This is the single place that
// starts a persistent, per-device login -- every entry point that
// authenticates a user (password login, register, Google OAuth) should call
// this instead of generating tokens directly, so every issued token pair is
// always backed by a session that can later be individually revoked
// (logout, password reset, account deletion).
type IssueTokensArgument struct {
	AccountId  string
	Email      string
	DeviceId   string
	DeviceName string
	UserAgent  string
	IpAddress  string
	// RequestedOrganizationId selects the active org for the new tokens
	// (empty means "use the account's primary org"). Used by the
	// "switch active organization" flow, which re-issues tokens for the
	// same session with a different active org.
	RequestedOrganizationId string
}

type IssueTokensResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

func IssueTokens(ctx context.Context, argument IssueTokensArgument) (res IssueTokensResult, err error) {
	sessionResult, err := sessions.Create(ctx, sessionspg.CreateArgument{
		AccountId: argument.AccountId,
		// RefreshTokenHash is filled in below once the refresh token exists
		// -- Create requires a non-empty placeholder since a session must
		// never be persisted with an empty hash (that would match nothing,
		// but is safer to just avoid entirely).
		RefreshTokenHash: "pending",
		DeviceId:         argument.DeviceId,
		DeviceName:       argument.DeviceName,
		UserAgent:        argument.UserAgent,
		IpAddress:        argument.IpAddress,
		ExpiresAt:        time.Now().Add(RefreshTokenTTL),
	})
	if err != nil {
		return res, err
	}

	accessToken, err := GenerateAccessToken(ctx, argument.AccountId, argument.Email, sessionResult.Session.Id, argument.RequestedOrganizationId)
	if err != nil {
		return res, err
	}

	refreshToken, err := GenerateRefreshToken(ctx, argument.AccountId, argument.Email, sessionResult.Session.Id, argument.RequestedOrganizationId)
	if err != nil {
		return res, err
	}

	hash := HashToken(refreshToken)
	if _, err = sessions.Update(ctx, sessionspg.UpdateArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "id", Operator: "=", Value: sessionResult.Session.Id},
					},
				},
			},
		},
		RefreshTokenHash: &hash,
	}); err != nil {
		return res, err
	}

	return IssueTokensResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(AccessTokenTTL),
	}, nil
}

// ReissueForSession re-mints an access+refresh token pair for an *existing*
// session (same device, same Session row -- its hash/expiry are updated in
// place) with a different active organization. This backs "switch active
// organization": it is a claim change, not a new login, so it must not
// create a second session row for what is still one device.
func ReissueForSession(ctx context.Context, sessionId, accountId, email, requestedOrganizationId string) (res IssueTokensResult, err error) {
	accessToken, err := GenerateAccessToken(ctx, accountId, email, sessionId, requestedOrganizationId)
	if err != nil {
		return res, err
	}

	refreshToken, err := GenerateRefreshToken(ctx, accountId, email, sessionId, requestedOrganizationId)
	if err != nil {
		return res, err
	}

	hash := HashToken(refreshToken)
	newExpiresAt := time.Now().Add(RefreshTokenTTL)
	now := time.Now()

	if _, err = sessions.Update(ctx, sessionspg.UpdateArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: sessionId}}},
			},
		},
		RefreshTokenHash: &hash,
		ExpiresAt:        &newExpiresAt,
		LastUsedAt:       &now,
	}); err != nil {
		return res, err
	}

	return IssueTokensResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(AccessTokenTTL),
	}, nil
}
