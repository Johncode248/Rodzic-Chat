package organizationspg

// BillingResponse is everything a director needs to pay (or check) the
// platform subscription: the fixed bank account, this organization's own
// generated transfer-title reference, and the price computed from its
// current ChildrenLimit tier (see pricing.go).
type BillingResponse struct {
	ChildrenLimit            int    `json:"children_limit"`
	MonthlyPriceGrosze       int64  `json:"monthly_price_grosze"`
	PaymentReference         string `json:"payment_reference"`
	SubscriptionFreeOverride bool   `json:"subscription_free_override"`
	BankAccountNumber        string `json:"bank_account_number"`
	BankAccountHolder        string `json:"bank_account_holder"`
}
