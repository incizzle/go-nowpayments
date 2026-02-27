package ipn

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/json"
	"fmt"

	"github.com/CIDgravity/go-nowpayments/config"
	"github.com/rotisserie/eris"
)

type IPNPaymentFees struct {
	Currency      string  `json:"currency"`
	DepositFee    float64 `json:"depositFee"`
	ServiceFee    float64 `json:"serviceFee"`
	WithdrawalFee float64 `json:"withdrawalFee"`
}

// PaymentStatus holds payment status related information
// Docs found on https://documenter.getpostman.com/view/7907941/2s93JusNJt#62a6d281-478d-4927-8cd0-f96d677b8de6
// Docs said IPN response is similar to PaymentStatus, but it's not the case
// And because this struct is used to compare a signature (callback), must be exactly the same
// Signature will be verified using VerifyRequestSignature method
// Fields are sorted alphabetically by JSON key to ensure consistent HMAC signature generation
type IPNPaymentStatus struct {
	ActuallyPaid       float64        `json:"actually_paid"`
	ActuallyPaidAtFiat float64        `json:"actually_paid_at_fiat"`
	BurningPercent     int            `json:"burning_percent"`
	CreatedAt          string         `json:"created_at"`
	Fee                IPNPaymentFees `json:"fee"`
	InvoiceID          int64          `json:"invoice_id"`
	Network            string         `json:"network"`
	NetworkPrecision   int            `json:"network_precision"`
	OrderDescription   string         `json:"order_description"`
	OrderID            string         `json:"order_id"`
	OutcomeAmount      float64        `json:"outcome_amount"`
	OutcomeCurrency    string         `json:"outcome_currency"`
	ParentPaymentId    *int64         `json:"parent_payment_id"`
	PayAddress         string         `json:"pay_address"`
	PayAmount          float64        `json:"pay_amount"`
	PayCurrency        string         `json:"pay_currency"`
	PayinExtraID       *int64         `json:"payin_extra_id"`
	PaymentExtraIds    []int64        `json:"payment_extra_ids"`
	PaymentID          int64          `json:"payment_id"`
	PaymentStatus      string         `json:"payment_status"`
	PriceAmount        float64        `json:"price_amount"`
	PriceCurrency      string         `json:"price_currency"`
	PurchaseID         string         `json:"purchase_id"`
	SmartContract      string         `json:"smart_contract"`
	UpdatedAt          string         `json:"updated_at"`
}

func VerifyRequestSignature(expectedSignature string, ipnNotificationBody IPNPaymentStatus) error {
	responseBodyAsBytes, err := json.Marshal(ipnNotificationBody)

	if err != nil {
		return err
	}

	// Create hmac sha512 using IPNSecretKey from config and response body
	// the sort of keys in struct can be important, sometimes the signatures will differ if order differ
	digest := hmac.New(sha512.New, []byte(config.IPNSecretKey()))
	digest.Write(responseBodyAsBytes)
	generatedSignature := digest.Sum(nil)

	// Compare generated signature to expectedSignature
	if fmt.Sprintf("%x", generatedSignature) == expectedSignature {
		return nil
	} else {
		return eris.Wrap(fmt.Errorf("HMAC signature does not match"), "IPN signature verification")
	}
}
