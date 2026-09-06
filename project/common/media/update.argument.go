package mediapg

import (
	"errors"
	"fmt"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/go-playground/validator/v10"
)

type UpdateArgument struct {
	Query   databasepg.Query `json:"query,omitempty" validate:"required"`
	Caption *string          `json:"caption,omitempty"`
}

func (a UpdateArgument) Populate(media Media) Media {

	if a.Caption != nil {
		media.Caption = *a.Caption
	}

	return media
}

func (a UpdateArgument) Validate() error {

	if err := validator.New().Struct(a); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
		}
	}

	return nil
}
