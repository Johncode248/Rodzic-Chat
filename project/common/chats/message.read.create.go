package chatspg

import (
	"errors"
	"fmt"
	"time"

	"github.com/Johncode248/Rodzic-Chat/library/expressions"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type MessageReadCreateArgument struct {
	Id        string `json:"id,omitempty"`
	MessageId string `json:"message_id,omitempty" validate:"required"`
	AccountId string `json:"account_id,omitempty" validate:"required"`
}

func (a MessageReadCreateArgument) Validate() error {
	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}
	return nil
}

func (a MessageReadCreateArgument) Populate() (read MessageRead) {
	read.Id = expressions.Select(a.Id != "", a.Id, uuid.NewString())
	read.MessageId = a.MessageId
	read.AccountId = a.AccountId
	read.ReadAt = time.Now()
	return read
}

type MessageReadCreateRequest struct {
	MessageId string `json:"message_id,omitempty"`
}

type MessageReadCreateResponse struct {
	MessageRead MessageRead `json:"message_read,omitempty"`
}
