package payments

import (
	"fmt"
	"net/url"

	"github.com/CIDgravity/go-nowpayments/config"
	"github.com/CIDgravity/go-nowpayments/core"
	"github.com/rotisserie/eris"
)

// ListByInvoiceOption are options applying to the list of payments by invoice
type ListByInvoiceOption struct {
	Limit    int
	Page     int
	DateFrom string
	DateTo   string
	OrderBy  string
	SortBy   string
}

// ListByInvoice returns a list of payments associated with the given invoice ID.
// This endpoint uses the Account API (account-api.nowpayments.io).
// JWT is required for this request.
func ListByInvoice(invoiceID string, o *ListByInvoiceOption) ([]*Payment[int64], error) {
	if invoiceID == "" {
		return nil, eris.New("empty invoice ID")
	}

	u := url.Values{}
	if o != nil {
		if o.Limit != 0 {
			u.Set("limit", fmt.Sprintf("%d", o.Limit))
		}
		if o.Page != 0 {
			u.Set("page", fmt.Sprintf("%d", o.Page))
		}
		if o.DateFrom != "" {
			u.Set("dateFrom", o.DateFrom)
		}
		if o.DateTo != "" {
			u.Set("dateTo", o.DateTo)
		}
		if o.SortBy != "" {
			u.Set("sortBy", o.SortBy)
		}
		if o.OrderBy != "" {
			u.Set("orderBy", o.OrderBy)
		}
	}

	tok, err := core.Authenticate(config.Login(), config.Password(), core.AccountAPIAuthURL())
	if err != nil {
		return nil, eris.Wrap(err, "list-by-invoice")
	}

	type plist struct {
		Data []*Payment[int64] `json:"data"`
	}

	pl := &plist{Data: make([]*Payment[int64], 0)}
	par := &core.SendParams{
		RouteName:       "payments-by-invoice",
		Path:            invoiceID,
		Into:            pl,
		Values:          u,
		JWTToken:        tok,
		BaseURLOverride: core.AccountAPIBaseURL(),
	}

	err = core.HTTPSend(par)
	if err != nil {
		return nil, err
	}

	return pl.Data, nil
}
