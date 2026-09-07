package paymentspg

// PaymentStatus is the "niezapłacone / zapłacone / zaksięgowane" lifecycle
// from product direction. Unpaid -> Paid normally happens automatically
// (the Stripe webhook, on a successful online checkout) but staff can also
// set any status directly at any time (PATCH /private/payments) -- that's
// what makes a bank-transfer payment (which Stripe never sees) reconcilable
// at all, and it's a deliberately simple/trusting model: a v1 tool for a
// single trusted director, not a workflow with approval gates.
type PaymentStatus int

const (
	PaymentStatusUnknown PaymentStatus = iota
	PaymentStatusUnpaid
	PaymentStatusPaid
	PaymentStatusReconciled
)

func (s PaymentStatus) String() string {
	switch s {
	case PaymentStatusUnpaid:
		return "Unpaid"
	case PaymentStatusPaid:
		return "Paid"
	case PaymentStatusReconciled:
		return "Reconciled"
	default:
		return "Unknown"
	}
}

func (s PaymentStatus) Ordinal() int {
	return int(s)
}

func MapPaymentStatus(value string) PaymentStatus {
	switch value {
	case "unpaid":
		return PaymentStatusUnpaid
	case "paid":
		return PaymentStatusPaid
	case "reconciled":
		return PaymentStatusReconciled
	default:
		return PaymentStatusUnknown
	}
}

// PaymentMethod records how a Payment actually got (or is being) paid --
// distinct from Status, since e.g. a bank-transfer payment can be marked
// Reconciled directly without ever going through Paid.
type PaymentMethod int

const (
	PaymentMethodUnset PaymentMethod = iota
	PaymentMethodOnline
	PaymentMethodBankTransfer
)

func (m PaymentMethod) String() string {
	switch m {
	case PaymentMethodOnline:
		return "Online"
	case PaymentMethodBankTransfer:
		return "BankTransfer"
	default:
		return "Unset"
	}
}

func (m PaymentMethod) Ordinal() int {
	return int(m)
}
