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
	Id        string `json:"id,omitempty"`
	ChannelId string `json:"channel_id,omitempty" validate:"required"`
	// OrganizationId is blank for an org-less parent-to-parent channel (see
	// project/service/transport/chats/parentchannel.message.create.go) --
	// not `validate:"required"` for that reason. Real authorization is the
	// explicit organization/membership check each transport handler does
	// before ever constructing this argument, not this struct tag.
	OrganizationId string      `json:"organization_id,omitempty"`
	SenderId       string      `json:"sender_id,omitempty" validate:"required"`
	Type           MessageType `json:"type,omitempty" validate:"required"`
	// max=1000: a chat message, not a document -- keeps storage/network payloads bounded and
	// discourages someone pasting a wall of text into what's meant to be a quick exchange.
	Content       string `json:"content,omitempty" validate:"max=1000"`
	AttachmentUrl string `json:"attachment_url,omitempty"`
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
