package payments

import (
	"context"
	"errors"
	"time"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	paymentspg "github.com/Johncode248/Rodzic-Chat/project/common/payments"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNoRecipients = errors.New("at least one recipient is required")
var ErrInvalidAmount = errors.New("base_amount_grosze must be greater than zero")

// Create makes one Payment row per recipient (see Payment's doc comment),
// all sharing a fresh BatchId and Category/Title/Description/DueDate, but
// each independently payable and trackable from here on.
func Create(ctx context.Context, organizationId, createdBy string, request paymentspg.CreateRequest) (res paymentspg.CreateResponse, err error) {
	ctx = logger.Enter(ctx, "payments.Create", organizationId, createdBy, request)
	defer logger.Exit(ctx, "payments.Create", &res, &err)

	if len(request.Recipients) == 0 {
		return res, ErrNoRecipients
	}
	if request.BaseAmountGrosze <= 0 {
		return res, ErrInvalidAmount
	}

	batchId := uuid.NewString()
	now := time.Now()
	category := paymentspg.MapPaymentCategory(request.Category)

	rows := make([]paymentspg.Payment, 0, len(request.Recipients))
	for _, recipient := range request.Recipients {
		rows = append(rows, paymentspg.Payment{
			Id:               uuid.NewString(),
			OrganizationId:   organizationId,
			BatchId:          batchId,
			ChildId:          recipient.ChildId,
			ParentId:         recipient.ParentId,
			CreatedBy:        createdBy,
			Category:         category,
			Title:            request.Title,
			Description:      request.Description,
			DueDate:          request.DueDate,
			BaseAmountGrosze: request.BaseAmountGrosze,
			PaymentMethod:    paymentspg.PaymentMethodUnset,
			Status:           paymentspg.PaymentStatusUnpaid,
			CreateTime:       now,
			UpdateTime:       now,
		})
	}

	// One row per recipient, but they must land together or not at all --
	// without a real transaction here, a crash or DB error partway through
	// the loop (e.g. assigning a payment to 3 parents) left some parents
	// charged and others silently not, with nothing to detect or undo it.
	if err = gormdb.GormDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range rows {
			if err := tx.Create(&rows[i]).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return res, err
	}

	res.Payments = rows

	return res, nil
}
