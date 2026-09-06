package chatspg

import (
	"errors"
	"fmt"
	"time"

	"github.com/Johncode248/Rodzic-Chat/library/expressions"
	"github.com/Johncode248/Rodzic-Chat/library/mutations"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type MessageCreateArgument struct {
	Id             string      `json:"id,omitempty"`
	ChannelId      string      `json:"channel_id,omitempty" validate:"required"`
	OrganizationId string      `json:"organization_id,omitempty" validate:"required"`
	SenderId       string      `json:"sender_id,omitempty" validate:"required"`
	Type           MessageType `json:"type,omitempty" validate:"required"`
	Content        string      `json:"content,omitempty"`
	AttachmentUrl  string      `json:"attachment_url,omitempty"`
}

func (a MessageCreateArgument) Validate() error {
	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}
	return nil
}

func (a MessageCreateArgument) Populate() (message Message) {
	message.Id = expressions.Select(a.Id != "", a.Id, uuid.NewString())
	message.ChannelId = a.ChannelId
	message.OrganizationId = a.OrganizationId
	message.SenderId = a.SenderId
	message.Type = a.Type
	message.Content = a.Content
	message.AttachmentUrl = a.AttachmentUrl
	message.CreateTime = time.Now()
	return message
}

type MessageCreateRequest struct {
	ChannelId     string `json:"channel_id,omitempty"`
	Content       string `json:"content,omitempty"`
	AttachmentUrl string `json:"attachment_url,omitempty"`
}

type MessageCreateResponse struct {
	Message Message `json:"message,omitempty"`
}

type MessageCreateMutation mutations.Transitional[MessageCreateArgument, MessageTransient]
