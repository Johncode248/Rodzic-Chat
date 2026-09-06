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

type ChannelCreateArgument struct {
	Id             string      `json:"id,omitempty"`
	OrganizationId string      `json:"organization_id,omitempty" validate:"required"`
	GroupId        string      `json:"group_id,omitempty"`
	Type           ChannelType `json:"type,omitempty" validate:"required"`
	Name           string      `json:"name,omitempty" validate:"required,max=128"`
}

func (a ChannelCreateArgument) Validate() error {
	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}
	return nil
}

func (a ChannelCreateArgument) Populate() (channel Channel) {
	channel.Id = expressions.Select(a.Id != "", a.Id, uuid.NewString())
	channel.OrganizationId = a.OrganizationId
	channel.GroupId = a.GroupId
	channel.Type = a.Type
	channel.Name = a.Name
	channel.CreateTime = time.Now()
	channel.UpdateTime = time.Now()
	return channel
}

type ChannelCreateRequest struct {
	GroupId string      `json:"group_id,omitempty"`
	Type    ChannelType `json:"type,omitempty"`
	Name    string      `json:"name,omitempty"`
}

type ChannelCreateResponse struct {
	Channel Channel `json:"channel,omitempty"`
}

type ChannelCreateMutation mutations.Transitional[ChannelCreateArgument, ChannelTransient]
