package sessionspg

import (
	"errors"
	"fmt"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/go-playground/validator/v10"
)

type UpdateArgument struct {
	Query            databasepg.Query `json:"query,omitempty" validate:"required"`
	RefreshTokenHash *string          `json:"-"`
	ExpiresAt        *time.Time       `json:"expires_at,omitempty"`
	LastUsedAt       *time.Time       `json:"last_used_at,omitempty"`
	RevokedAt        *time.Time       `json:"revoked_at,omitempty"`
}

func (a UpdateArgument) Populate(session Session) Session {

	session.UpdateTime = time.Now()

	if a.RefreshTokenHash != nil {
		session.RefreshTokenHash = *a.RefreshTokenHash
	}

	if a.ExpiresAt != nil {
		session.ExpiresAt = *a.ExpiresAt
	}

	if a.LastUsedAt != nil {
		session.LastUsedAt = *a.LastUsedAt
	}

	if a.RevokedAt != nil {
		session.RevokedAt = a.RevokedAt
	}

	return session
}

func (a UpdateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}
