package paymentspg

import "time"

// Payment is a one-time charge a director/co-director assigns to one
// parent, for one child. "Assigning to several parents at once" (product
// direction) creates one Payment row per parent, all sharing BatchId, so
// each is independently payable/trackable rather than a shared split.
//
// Amounts are grosze (1 zł = 100 grosze) to avoid float rounding.
// BaseAmountGrosze is fixed at creation and is exactly what the school
// receives, however the parent pays -- PlatformFeeGrosze,
// StripeFeeEstimateGrosze and TotalAmountGrosze are all zero until the
// parent actually starts an online checkout (see
// project/service/internal/payments.ComputeCheckoutAmounts), since a
// bank-transfer payment never needs them at all (product direction: no
// platform/Stripe fee applies outside the app).
type Payment struct {
	Id             string          `json:"id,omitempty" gorm:"primaryKey"`
	OrganizationId string          `json:"organization_id,omitempty" gorm:"not null;index"`
	BatchId        string          `json:"batch_id,omitempty" gorm:"index"`
	ChildId        string          `json:"child_id,omitempty" gorm:"not null;index"`
	ParentId       string          `json:"parent_id,omitempty" gorm:"not null;index"`
	CreatedBy      string          `json:"created_by,omitempty" gorm:"not null"`
	Category       PaymentCategory `json:"category,omitempty"`
	Title          string          `json:"title,omitempty" gorm:"not null"`
	Description    string          `json:"description,omitempty"`
	DueDate        *time.Time      `json:"due_date,omitempty"`

	BaseAmountGrosze        int64 `json:"base_amount_grosze,omitempty" gorm:"not null"`
	PlatformFeeGrosze       int64 `json:"platform_fee_grosze,omitempty"`
	StripeFeeEstimateGrosze int64 `json:"stripe_fee_estimate_grosze,omitempty"`
	TotalAmountGrosze       int64 `json:"total_amount_grosze,omitempty"`
	ActualStripeFeeGrosze   int64 `json:"actual_stripe_fee_grosze,omitempty"`

	PaymentMethod PaymentMethod `json:"payment_method,omitempty"`
	Status        PaymentStatus `json:"status,omitempty" gorm:"not null"`

	StripeCheckoutSessionId string `json:"stripe_checkout_session_id,omitempty"`
	StripePaymentIntentId   string `json:"stripe_payment_intent_id,omitempty"`

	PaidAt       *time.Time `json:"paid_at,omitempty"`
	ReconciledAt *time.Time `json:"reconciled_at,omitempty"`
	ReconciledBy string     `json:"reconciled_by,omitempty"`

	CreateTime time.Time `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime time.Time `json:"update_time,omitempty" gorm:"not null"`
}

// PaymentRecipient is one {child, parent} pair in a CreateRequest -- see
// Payment's doc comment on why one Payment row is created per recipient.
type PaymentRecipient struct {
	ChildId  string `json:"child_id,omitempty"`
	ParentId string `json:"parent_id,omitempty"`
}

type CreateRequest struct {
	Category         string             `json:"category,omitempty"`
	Title            string             `json:"title,omitempty"`
	Description      string             `json:"description,omitempty"`
	BaseAmountGrosze int64              `json:"base_amount_grosze,omitempty"`
	DueDate          *time.Time         `json:"due_date,omitempty"`
	Recipients       []PaymentRecipient `json:"recipients,omitempty"`
}

type CreateResponse struct {
	Payments []Payment `json:"payments,omitempty"`
}

type SelectResponse struct {
	Payments []Payment `json:"payments,omitempty"`
}

// UpdateRequest fields are only settable while the payment is still Unpaid
// (Title/Description/BaseAmountGrosze/DueDate) except Status, which staff
// can change at any time -- see PaymentStatus's doc comment.
type UpdateRequest struct {
	Status           *string    `json:"status,omitempty"`
	Title            *string    `json:"title,omitempty"`
	Description      *string    `json:"description,omitempty"`
	BaseAmountGrosze *int64     `json:"base_amount_grosze,omitempty"`
	DueDate          *time.Time `json:"due_date,omitempty"`
}

type UpdateResponse struct {
	Payment Payment `json:"payment,omitempty"`
}

type DeleteResponse struct {
	Payments []Payment `json:"payments,omitempty"`
}

// CheckoutResponse is what POST /private/payments/checkout?id= returns: a
// Stripe Checkout URL the parent's client should open.
type CheckoutResponse struct {
	Url               string `json:"url,omitempty"`
	TotalAmountGrosze int64  `json:"total_amount_grosze,omitempty"`
}
