package payments

import (
	"errors"
	"net/http"
	"testing"

	"github.com/CIDgravity/go-nowpayments/core"
	"github.com/CIDgravity/go-nowpayments/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStatus(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name      string
		paymentID string
		init      func(*mocks.HTTPClient)
		after     func(*mocks.HTTPClient, *PaymentStatus, error)
	}{
		{"empty payment ID", "", nil,
			func(c *mocks.HTTPClient, s *PaymentStatus, err error) {
				require.Error(t, err)
				assert.Nil(s)
			},
		},
		{"ok", "PID",
			func(c *mocks.HTTPClient) {
				resp := newResponseOK(`{"payment_status":"done","pay_amount":10.0,"outcome_amount":9.5,"outcome_currency":"btc","burning_percent":1}`)
				c.EXPECT().Do(mock.Anything).Run(func(r *http.Request) {
					assert.Equal("/v1/payment/PID", r.URL.Path, "bad endpoint")
				}).Return(resp, nil)
			},
			func(c *mocks.HTTPClient, s *PaymentStatus, err error) {
				assert.NoError(err)
				assert.NotNil(s)
				assert.Equal(10.0, s.PayAmount)
				assert.Equal("done", s.Status)
				assert.Equal(9.5, s.OutcomeAmount)
				assert.Equal("btc", s.OutcomeCurrency)
				assert.Equal(1, s.BurningPercent)
				c.AssertNumberOfCalls(t, "Do", 1)
			},
		},
		{"api error", "ID",
			func(c *mocks.HTTPClient) {
				c.EXPECT().Do(mock.Anything).Return(nil, errors.New("network error"))
			},
			func(c *mocks.HTTPClient, s *PaymentStatus, err error) {
				assert.Error(err)
				assert.Nil(s)
				assert.Equal("payment-status: network error", err.Error())
				c.AssertNumberOfCalls(t, "Do", 1)
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
			got, err := Status(tt.paymentID)
			if tt.after != nil {
				tt.after(c, got, err)
			}
		})
	}
}
