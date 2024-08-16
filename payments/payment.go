package payments

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/CIDgravity/go-nowpayments/core"
	"github.com/rotisserie/eris"
)

// PaymentAmount defines common fields used in PaymentArgs and Payment structs
type PaymentAmount struct {
	PriceAmount      float64 `json:"price_amount"`
	PriceCurrency    string  `json:"price_currency"`
	PayCurrency      string  `json:"pay_currency"`
	PayAmount        string  `json:"pay_amount,omitempty"`
	CallbackURL      string  `json:"ipn_callback_url,omitempty"`
	OrderID          string  `json:"order_id,omitempty"`
	OrderDescription string  `json:"order_description,omitempty"`
}

// PaymentArgs are the arguments used to make a payment
type PaymentArgs struct {
	PaymentAmount

	// FeePaidByUser is optional, required for fixed-rate exchanges with all fees paid by users.
	FeePaidByUser bool `json:"is_fee_paid_by_user,omitempty"`
	// FixedRate is optional, required for fixed-rate exchanges.
	FixedRate bool `json:"fixed_rate,omitempty"`
	// PayoutAddress is optional, usually the funds will go to the address you specify in
	// your personal account. In case you want to receive funds on another address, you can specify
	// it in this parameter.
	PayoutAddress string `json:"payout_address,omitempty"`
	// PayAmount is optional, the amount that users have to pay for the order stated in crypto.
	// You can either specify it yourself, or we will automatically convert the amount indicated
	// in price_amount.
	PayAmount float64 `json:"pay_amount,omitempty"`
	// PayoutCurrency for the cryptocurrency name.
	PayoutCurrency string `json:"payout_currency,omitempty"`
	// PayoutExtraID is optional, extra id or memo or tag for external payout_address.
	PayoutExtraID string `json:"payout_extra_id,omitempty"`
	// PurchaseID is optional, id of purchase for which you want to create another
	// payment, only used for several payments for one order.
	PurchaseID string `json:"purchase_id,omitempty"`
	// optional, case which you want to test (sandbox only).
	Case string `json:"case,omitempty"`
}

// Payment holds payment related information once we get a response
// This struct will be used in multiple API calls
// Inconsistency on their side: we need to use generics, because some times payment_id is string or int64
type Payment[T string | int64] struct {
	PaymentAmount

	ID           T           `json:"payment_id"`
	InvoiceID    json.Number `json:"invoice_id"`
	Status       string      `json:"payment_status"`
	PayAddress   string      `json:"pay_address"`
	PayinExtraID string      `json:"payin_extra_id"`
	PayAmount    float64     `json:"pay_amount"`
	ActuallyPaid float64     `json:"actually_paid"`
	PayCurrency  string      `json:"pay_currency"`
	PurchaseID   json.Number `json:"purchase_id"`

	OutcomeAmount   float64 `json:"outcome_amount"`
	OutcomeCurrency string  `json:"outcome_currency"`

	PayoutHash *string `json:"payout_hash"`
	PayinHash  *string `json:"payin_hash"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`

	Type                   string  `json:"type"`
	AmountReceived         float64 `json:"amount_received"`
	BurningPercent         int     `json:"burning_percent"`
	ExpirationEstimateDate string  `json:"expiration_estimate_date,omitempty"`
	Network                string  `json:"network,omitempty"`
	NetworkPrecision       int     `json:"network_precision,omitempty"`
	SmartContract          string  `json:"smart_contract,omitempty"`
	TimeLimit              string  `json:"time_limit,omitempty"`
}

// New creates a payment
func New(pa *PaymentArgs) (*Payment[string], error) {
	if pa == nil {
		return nil, errors.New("nil payment args")
	}

	d, err := json.Marshal(pa)
	if err != nil {
		return nil, eris.Wrap(err, "payment args")
	}

	p := &Payment[string]{}

	par := &core.SendParams{
		RouteName: "payment-create",
		Into:      &p,
		Body:      strings.NewReader(string(d)),
	}

	err = core.HTTPSend(par)
	if err != nil {
		return nil, err
	}

	return p, nil
}

type InvoicePaymentArgs struct {
	InvoiceID        string `json:"iid"`
	PayCurrency      string `json:"pay_currency"`
	PurchaseID       string `json:"purchase_id,omitempty"`
	OrderDescription string `json:"order_description,omitempty"`
	CustomerEmail    string `json:"customer_email,omitempty"`
	PayoutCurrency   string `json:"payout_currency,omitempty"`
	PayoutExtraID    string `json:"payout_extra_id,omitempty"`
	PayoutAddress    string `json:"payout_address,omitempty"`
}

// NewFromInvoice creates a payment from an existing invoice. ID is the invoice's identifier.
func NewFromInvoice(ipa *InvoicePaymentArgs) (*Payment[string], error) {
	if ipa == nil {
		return nil, errors.New("nil invoice payment args")
	}

	d, err := json.Marshal(ipa)
	if err != nil {
		return nil, eris.Wrap(err, "payment from invoice args")
	}

	p := &Payment[string]{}

	par := &core.SendParams{
		RouteName: "invoice-payment",
		Into:      &p,
		Body:      strings.NewReader(string(d)),
	}

	err = core.HTTPSend(par)
	if err != nil {
		return nil, err
	}

	return p, nil
}
