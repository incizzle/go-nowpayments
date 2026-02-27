package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/CIDgravity/go-nowpayments/config"
	"github.com/rotisserie/eris"
)

// BaseURL is the URL to NOWPayment's service
type BaseURL string

const (
	ProductionBaseURL           BaseURL = "https://api.nowpayments.io/v1"
	SandBoxBaseURL                      = "https://api-sandbox.nowpayments.io/v1"
	AccountAPIProductionURL     BaseURL = "https://account-api.nowpayments.io"
	AccountAPISandBoxURL        BaseURL = "https://account-api-sandbox.nowpayments.io"
	AccountAPIProductionAuthURL BaseURL = "https://account-api.nowpayments.io/v1"
	AccountAPISandBoxAuthURL    BaseURL = "https://account-api-sandbox.nowpayments.io/v1"
)

// SendParams are parameters needed to build and send an HTTP request to the service
type SendParams struct {
	Body            io.Reader
	Into            interface{}
	Path            string
	RouteName       string
	Values          url.Values
	JWTToken        string
	BaseURLOverride BaseURL // optional: override the default base URL for this request
}

type routeAttr struct {
	method string
	path   string
}

// V2ResponseFormat handle some inconsistency on their side
// some response are at root level, sometimes in data (list only) and sometimes under result key
// because of many result format, we use generics to instantiate the struct with the correct response format
type V2ResponseFormat[T interface{}] struct {
	Result T `json:"result"`
}

var routes map[string]routeAttr = map[string]routeAttr{
	"auth":   {http.MethodPost, "/auth"},
	"status": {http.MethodGet, "/status"},

	// Currencies and estimation routes
	"currencies": {http.MethodGet, "/currencies"},
	"estimate":   {http.MethodGet, "/estimate"},

	// Payments routes
	"invoice-create":      {http.MethodPost, "/invoice"},
	"invoice-payment":     {http.MethodPost, "/invoice-payment"},
	"last-estimate":       {http.MethodPost, "/payment"},
	"min-amount":          {http.MethodGet, "/min-amount"},
	"payment-create":      {http.MethodPost, "/payment"},
	"payment-status":      {http.MethodGet, "/payment"},
	"payments-list":       {http.MethodGet, "/payment/"},
	"payments-by-invoice": {http.MethodGet, "/payments/by-invoice"},
	"selected-currencies": {http.MethodGet, "/merchant/coins"},

	// Subscription routes
	"subscription-create":       {http.MethodPost, "/subscriptions/plans"},
	"subscription-update":       {http.MethodPatch, "/subscriptions/plans"},
	"subscription-single":       {http.MethodGet, "/subscriptions/plans"},
	"subscription-list":         {http.MethodGet, "/subscriptions/plans"},
	"subscription-create-email": {http.MethodPost, "/subscriptions"},

	// Recurring payments routes
	"recurring-payment-create": {http.MethodPost, "/subscriptions"},
	"recurring-payment-single": {http.MethodGet, "/subscriptions"},
	"recurring-payment-list":   {http.MethodGet, "/subscriptions"},
	"recurring-payment-delete": {http.MethodDelete, "/subscriptions"},

	// Custody routes
	"custody-create-account":       {http.MethodPost, "/sub-partner/balance"},
	"custody-account-balance":      {http.MethodGet, "/sub-partner/balance"},
	"custody-transfer-create":      {http.MethodPost, "/sub-partner/transfer"},
	"custody-list-transfers":       {http.MethodGet, "/sub-partner/transfers"},
	"custody-transfer-single":      {http.MethodGet, "/sub-partner/transfer"},
	"custody-list-users":           {http.MethodGet, "/sub-partner"},
	"custody-deposit-with-payment": {http.MethodPost, "/sub-partner/payment"},
	"custody-deposit-from-master":  {http.MethodPost, "/sub-partner/deposit"},
	"custody-payment-list":         {http.MethodGet, "/sub-partner/payments"},
	"custody-write-off-to-master":  {http.MethodPost, "/sub-partner/write-off"},
}

var (
	defaultURL BaseURL = SandBoxBaseURL
)

var debug = false

// WithDebug prints out debugging info about HTTP traffic
func WithDebug(d bool) {
	debug = d
}

// UseBaseURL sets the base URL to use to connect to NOWPayment's API
func UseBaseURL(b BaseURL) {
	defaultURL = b
}

// AccountAPIBaseURL returns the Account API base URL corresponding to the current default URL.
// If the default is sandbox, it returns the sandbox account API URL; otherwise production.
func AccountAPIBaseURL() BaseURL {
	if defaultURL == SandBoxBaseURL {
		return AccountAPISandBoxURL
	}
	return AccountAPIProductionURL
}

// AccountAPIAuthURL returns the Account API auth base URL (with /v1 path) corresponding to the current default URL.
func AccountAPIAuthURL() BaseURL {
	if defaultURL == SandBoxBaseURL {
		return AccountAPISandBoxAuthURL
	}
	return AccountAPIProductionAuthURL
}

// HTTPSend sends to endpoint with an optional request body and get the HTTP response result in into
func HTTPSend(p *SendParams) error {
	if p == nil {
		return eris.New("nil params")
	}

	method, path := routes[p.RouteName].method, routes[p.RouteName].path
	if path == "" {
		return eris.New(fmt.Sprintf("bad route name: empty path for endpoint %q", p.RouteName))
	}

	base := defaultURL
	if p.BaseURLOverride != "" {
		base = p.BaseURLOverride
	}
	u := string(base) + path
	if p.Path != "" {
		u += "/" + p.Path
	}

	if p.Values != nil {
		u += "?" + p.Values.Encode()
	}

	req, err := http.NewRequest(method, u, p.Body)
	if err != nil {
		return eris.Wrap(err, p.RouteName)
	}

	// Extra headers
	req.Header.Add("X-API-KEY", config.APIKey())
	if p.Body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	if p.JWTToken != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", p.JWTToken))
	}

	// Debug mode
	if debug {
		fmt.Println(">>> DEBUG REQUEST")
		fmt.Printf("X-API-KEY: %s\n", req.Header.Get("X-API-KEY"))
		fmt.Printf("Authorization: %s\n", req.Header.Get("Authorization"))
		fmt.Println(req.Method, req.URL.String())
		fmt.Println("<<< END DEBUG REQUEST")
	}

	res, err := client.Do(req)
	if err != nil {
		return eris.Wrap(err, p.RouteName)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		if debug {
			fmt.Printf(">>> DEBUG HTTP error %d: %s\n", res.StatusCode, res.Status)
		}

		type errResp struct {
			StatusCode int    `json:"statusCode"`
			Code       string `json:"code"`
			Message    string `json:"message"`
		}

		z := &errResp{}
		d := json.NewDecoder(res.Body)

		err = d.Decode(&z)
		if err != nil {
			return eris.Wrapf(err, "%s: JSON decode error", p.RouteName)
		}

		return eris.New(fmt.Sprintf("code %d (%s): %s", z.StatusCode, z.Code, z.Message))
	}

	if debug {
		fmt.Println(">>> DEBUG RAW RESPONSE BODY")

		all, err := io.ReadAll(res.Body)
		if err != nil {
			return eris.Wrap(err, "debug response")
		}

		fmt.Println(string(all))
		fmt.Println("<<< END DEBUG RAW RESPONSE BODY")
		return eris.Wrap(json.Unmarshal(all, &p.Into), p.RouteName)
	}

	d := json.NewDecoder(res.Body)
	err = d.Decode(&p.Into)
	return eris.Wrap(err, p.RouteName)
}
