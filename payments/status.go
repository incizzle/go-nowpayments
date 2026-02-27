package payments

import (
	"github.com/CIDgravity/go-nowpayments/core"
	"github.com/rotisserie/eris"
)

// PaymentStatus is the actual information about a payment
type PaymentStatus struct {
	ID              int64   `json:"payment_id"`
	ParentPaymentID *int64  `json:"parent_payment_id"`
	InvoiceID       int64   `json:"invoice_id"`
	Status          string  `json:"payment_status"`
	PayAddress      string  `json:"pay_address"`
	PayinExtraID    string  `json:"payin_extra_id"`
	PriceAmount     float64 `json:"price_amount"`
	PriceCurrency   string  `json:"price_currency"`
	PayAmount       float64 `json:"pay_amount"`
	ActuallyPaid    float64 `json:"actually_paid"`
	PayCurrency     string  `json:"pay_currency"`
	OrderID         string  `json:"order_id"`
	OrderDescription string `json:"order_description"`
	PurchaseID      int64   `json:"purchase_id"`
	OutcomeAmount   float64 `json:"outcome_amount"`
	OutcomeCurrency string  `json:"outcome_currency"`
	PayoutHash      *string `json:"payout_hash"`
	PayinHash       *string `json:"payin_hash"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
	Type            string  `json:"type"`
	AmountReceived  float64 `json:"amount_received"`
	BurningPercent  int     `json:"burning_percent"`
	Network         string  `json:"network,omitempty"`
	NetworkPrecision int    `json:"network_precision,omitempty"`
	SmartContract   string  `json:"smart_contract,omitempty"`
	ExpirationEstimateDate string `json:"expiration_estimate_date,omitempty"`
	TimeLimit       string  `json:"time_limit,omitempty"`
}

// Status gets the actual information about the payment. You need to provide the payment ID.
func Status(paymentID string) (*PaymentStatus, error) {
	if paymentID == "" {
		return nil, eris.New("empty payment ID")
	}

	st := &PaymentStatus{}
	par := &core.SendParams{
		RouteName: "payment-status",
		Path:      paymentID,
		Into:      &st,
	}

	err := core.HTTPSend(par)
	if err != nil {
		return nil, err
	}

	return st, nil
}
