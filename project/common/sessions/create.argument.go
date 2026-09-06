package sessionspg

import (
	"errors"
	"fmt"
	"time"

	"github.com/Johncode248/Rodzic-Chat/library/expressions"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CreateArgument struct {
	Id               string    `json:"id,omitempty"`
	AccountId        string    `json:"account_id,omitempty" validate:"required"`
	RefreshTokenHash string    `json:"-" validate:"required"`
	DeviceId         string    `json:"device_id,omitempty"`
	DeviceName       string    `json:"device_name,omitempty"`
	UserAgent        string    `json:"user_agent,omitempty"`
	IpAddress        string    `json:"ip_address,omitempty"`
	ExpiresAt        time.Time `json:"expires_at,omitempty" validate:"required"`
}

func (a CreateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}

func (a CreateArgument) Populate() (session Session) {

	now := time.Now()

	session.Id = expressions.Select(a.Id != "", a.Id, uuid.NewString())
	session.AccountId = a.AccountId
	session.RefreshTokenHash = a.RefreshTokenHash
	session.DeviceId = a.DeviceId
	session.DeviceName = a.DeviceName
	session.UserAgent = a.UserAgent
	session.IpAddress = a.IpAddress
	session.IssuedAt = now
	session.ExpiresAt = a.ExpiresAt
	session.LastUsedAt = now
	session.CreateTime = now
	session.UpdateTime = now

	return session
}
