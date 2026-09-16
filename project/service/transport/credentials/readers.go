package credentials

import (
	"context"
	"errors"
)

func GetUserIdFromContext(ctx context.Context) (userId string, err error) {
	userId = ctx.Value("userId").(string)
	if userId == "" {
		return "", errors.New("userId not found in context")
	}

	return userId, nil
}

func GetUserEmailFromContext(ctx context.Context) (email string, err error) {
	email = ctx.Value("userEmail").(string)
	if email == "" {
		return "", errors.New("email not found in context")
	}

	return email, nil
}

func GetOrganizationIdFromContext(ctx context.Context) (organizationId string, err error) {
	organizationId = ctx.Value("organizationId").(string)
	if organizationId == "" {
		return "", errors.New("organizationId not found in context")
	}

	return organizationId, nil
}

func GetSessionIdFromContext(ctx context.Context) (sessionId string, err error) {
	sessionId, _ = ctx.Value("sessionId").(string)
	if sessionId == "" {
		return "", errors.New("sessionId not found in context")
	}

	return sessionId, nil
}

func GetAccountTypeFromContext(ctx context.Context) (accountType int, err error) {
	accountType, ok := ctx.Value("accountType").(int)
	if !ok {
		return 0, errors.New("accountType not found in context")
	}

	return accountType, nil
}
