package payments

import (
	"errors"
	"net/http"
	"testing"

	"github.com/CIDgravity/go-nowpayments/core"
	"github.com/CIDgravity/go-nowpayments/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListByInvoice(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name      string
		invoiceID string
		o         *ListByInvoiceOption
		init      func(*mocks.HTTPClient)
		after     func([]*Payment[int64], error)
	}{
		{"empty invoice ID", "", nil, nil,
			func(ps []*Payment[int64], err error) {
				assert.Nil(ps)
				assert.Error(err)
			},
		},
		{"route and response", "5463950329", nil,
			func(c *mocks.HTTPClient) {
				c.EXPECT().Do(mock.Anything).Call.Return(
					func(req *http.Request) *http.Response {
						switch req.URL.Path {
						case "/v1/auth":
							return newResponseOK(`{"token":"tok"}`)
						case "/payments/by-invoice/5463950329":
							return newResponseOK(`{"data":[{"payment_id":100}]}`)
						default:
							t.Fatalf("unexpected route call %q", req.URL.String())
						}
						return nil
					}, nil)
			},
			func(ps []*Payment[int64], err error) {
				assert.NoError(err)
				if assert.Len(ps, 1) {
					assert.Equal(int64(100), ps[0].ID)
				}
			},
		},
		{"auth fail", "123", nil,
			func(c *mocks.HTTPClient) {
				c.EXPECT().Do(mock.Anything).Return(nil, errors.New("bad credentials"))
			},
			func(ps []*Payment[int64], err error) {
				assert.Nil(ps)
				assert.Error(err)
				assert.Equal("list-by-invoice: auth: bad credentials", err.Error())
			},
		},
		{"with limit option", "5463950329", &ListByInvoiceOption{Limit: 10},
			func(c *mocks.HTTPClient) {
				c.EXPECT().Do(mock.Anything).Call.Return(
					func(req *http.Request) *http.Response {
						switch req.URL.Path {
						case "/v1/auth":
							return newResponseOK(`{"token":"tok"}`)
						case "/payments/by-invoice/5463950329":
							assert.Equal("limit=10", req.URL.RawQuery)
							return newResponseOK(`{"data":[{"payment_id":200}]}`)
						default:
							t.Fatalf("unexpected route call %q", req.URL.String())
						}
						return nil
					}, nil)
			},
			func(ps []*Payment[int64], err error) {
				assert.NoError(err)
				if assert.Len(ps, 1) {
					assert.Equal(int64(200), ps[0].ID)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := mocks.NewHTTPClient(t)
			core.UseClient(c)
			if tt.init != nil {
				tt.init(c)
			}
			got, err := ListByInvoice(tt.invoiceID, tt.o)
			if tt.after != nil {
				tt.after(got, err)
			}
		})
	}
}
