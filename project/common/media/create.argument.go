package mediapg

import (
	"errors"
	"fmt"
	"time"

	"github.com/Johncode248/Rodzic-Chat/library/expressions"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CreateArgument struct {
	Id               string `json:"id,omitempty"`
	OrganizationId   string `json:"organization_id,omitempty" validate:"required"`
	EventId          string `json:"event_id,omitempty"`
	GroupId          string `json:"group_id,omitempty"`
	UploaderId       string `json:"uploader_id,omitempty" validate:"required"`
	ObjectKey        string `json:"object_key,omitempty" validate:"required"`
	OriginalFilename string `json:"original_filename,omitempty"`
	ContentType      string `json:"content_type,omitempty"`
	Caption          string `json:"caption,omitempty"`
	Category         string `json:"category,omitempty"`
}

func (a CreateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}

func (a CreateArgument) Populate() (media Media) {

	media.Id = expressions.Select(a.Id != "", a.Id, uuid.NewString())
	media.OrganizationId = a.OrganizationId
	media.EventId = a.EventId
	media.GroupId = a.GroupId
	media.UploaderId = a.UploaderId
	media.ObjectKey = a.ObjectKey
	media.OriginalFilename = a.OriginalFilename
	media.ContentType = a.ContentType
	media.Caption = a.Caption
	media.Category = a.Category
	media.CreateTime = time.Now()

	return media
}
