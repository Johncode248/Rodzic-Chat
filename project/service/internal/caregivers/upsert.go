package caregivers

import (
	"context"
	"errors"
	"strings"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	caregiverspg "github.com/Johncode248/Rodzic-Chat/project/common/caregivers"
	"github.com/google/uuid"
)

var ErrInvalidType = errors.New("invalid caregiver type")
var ErrDisplayNameRequired = errors.New("display_name is required")
var ErrModeRequired = errors.New("choose at least one way of working: online, at your place or travelling to the client")
var ErrCityRequired = errors.New("city is required for in-person work")

// Upsert creates or updates the caller's own Caregiver profile for the
// given type -- there's at most one per (account, type). A parent who is
// herself a nanny AND a tutor gets two independent rows, one per type.
func Upsert(ctx context.Context, accountId string, req caregiverspg.UpsertRequest) (res caregiverspg.Caregiver, err error) {
	caregiverType := caregiverspg.CaregiverTypeFromOrdinal(req.Type)
	if caregiverType != caregiverspg.CaregiverTypeNanny && caregiverType != caregiverspg.CaregiverTypeTutor {
		return res, ErrInvalidType
	}
	if req.DisplayName == "" {
		return res, ErrDisplayNameRequired
	}

	if !req.Online && !req.Stationary && !req.Travel {
		return res, ErrModeRequired
	}
	if (req.Stationary || req.Travel) && strings.TrimSpace(req.City) == "" {
		return res, ErrCityRequired
	}

	now := time.Now()

	existing, err := getMine(ctx, accountId, caregiverType)
	if err != nil {
		return res, err
	}

	if existing != nil {
		updated := *existing
		updated.DisplayName = req.DisplayName
		updated.Description = req.Description
		updated.Experience = req.Experience
		updated.HourlyRate = req.HourlyRate
		updated.City = req.City
		updated.Phone = req.Phone
		updated.Email = req.Email
		applyOffer(&updated, req)
		// Any edit goes back to the admin queue.
		updated.ReviewStatus = caregiverspg.ReviewStatusPending
		updated.RejectReason = ""
		updated.ReviewedAt = nil
		updated.UpdateTime = now

		if err = gormdb.Update(ctx, updated, databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: updated.Id}}},
			},
		}); err != nil {
			return res, err
		}
		return updated, nil
	}

	created := caregiverspg.Caregiver{
		Id:           uuid.NewString(),
		AccountId:    accountId,
		Type:         caregiverType,
		DisplayName:  req.DisplayName,
		Description:  req.Description,
		Experience:   req.Experience,
		HourlyRate:   req.HourlyRate,
		City:         req.City,
		Phone:        req.Phone,
		Email:        req.Email,
		ReviewStatus: caregiverspg.ReviewStatusPending,
		CreateTime:   now,
		UpdateTime:   now,
	}
	applyOffer(&created, req)
	if err = gormdb.Create(ctx, &created); err != nil {
		return res, err
	}
	return created, nil
}

// applyOffer copies the "what and where" part of a request onto a profile.
func applyOffer(c *caregiverspg.Caregiver, req caregiverspg.UpsertRequest) {
	c.Subjects = strings.TrimSpace(req.Subjects)
	c.Online, c.Stationary, c.Travel = req.Online, req.Stationary, req.Travel
	c.ServiceCities = strings.TrimSpace(req.ServiceCities)
	c.TravelRadiusKm = req.TravelRadiusKm
	if !req.Travel {
		c.TravelRadiusKm = 0
	}
}
