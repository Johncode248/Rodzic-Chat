package chatspg

import (
	"errors"
	"fmt"
	"time"

	"github.com/Johncode248/Rodzic-Chat/library/expressions"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// PollCreateArgument is the top-level "create a poll" call: it creates a
// Message (Type=Poll) + Poll + PollOptions together in one step (see
// project/service/internal/chats/poll.create.go).
type PollCreateArgument struct {
	ChannelId     string     `json:"channel_id,omitempty" validate:"required"`
	Question      string     `json:"question,omitempty" validate:"required,max=512"`
	Options       []string   `json:"options,omitempty" validate:"required,min=2,dive,required,max=256"`
	AllowMultiple bool       `json:"allow_multiple,omitempty"`
	ClosesAt      *time.Time `json:"closes_at,omitempty"`

	// Populated by the transport layer from JWT context, never trusted from
	// the client body.
	OrganizationId string `json:"-"`
	SenderId       string `json:"-"`
}

func (a PollCreateArgument) Validate() error {
	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}
	return nil
}

type PollCreateRequest struct {
	ChannelId     string     `json:"channel_id,omitempty"`
	Question      string     `json:"question,omitempty"`
	Options       []string   `json:"options,omitempty"`
	AllowMultiple bool       `json:"allow_multiple,omitempty"`
	ClosesAt      *time.Time `json:"closes_at,omitempty"`
}

type PollCreateResponse struct {
	Message Message      `json:"message,omitempty"`
	Poll    Poll         `json:"poll,omitempty"`
	Options []PollOption `json:"options,omitempty"`
}

// PollRowCreateArgument creates the bare Poll row itself, once the owning
// Message already exists.
type PollRowCreateArgument struct {
	Id            string
	MessageId     string
	Question      string
	AllowMultiple bool
	ClosesAt      *time.Time
}

func (a PollRowCreateArgument) Populate() (poll Poll) {
	poll.Id = expressions.Select(a.Id != "", a.Id, uuid.NewString())
	poll.MessageId = a.MessageId
	poll.Question = a.Question
	poll.AllowMultiple = a.AllowMultiple
	poll.ClosesAt = a.ClosesAt
	poll.CreateTime = time.Now()
	return poll
}

// PollOptionCreateArgument creates a single PollOption row.
type PollOptionCreateArgument struct {
	Id         string
	PollId     string
	Text       string
	OrderIndex int
}

func (a PollOptionCreateArgument) Populate() (option PollOption) {
	option.Id = expressions.Select(a.Id != "", a.Id, uuid.NewString())
	option.PollId = a.PollId
	option.Text = a.Text
	option.OrderIndex = a.OrderIndex
	return option
}
