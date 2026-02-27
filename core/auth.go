package core

import (
	"fmt"
	"strings"
)

type token struct {
	Token string `json:"token"`
}

// Authenticate is used for obtaining a JWT token.
// JWT is required only for payout request API call.
// An optional baseURL can be provided to authenticate against a different host
// (e.g. the Account API). If not provided, the default base URL is used.
func Authenticate(email, password string, baseURL ...BaseURL) (string, error) {
	r := strings.NewReader(fmt.Sprintf(`{
			"email": "%s",
			"password": "%s"
		}`, email, password))

	t := &token{}

	par := &SendParams{
		RouteName: "auth",
		Body:      r,
		Into:      &t,
	}

	if len(baseURL) > 0 && baseURL[0] != "" {
		par.BaseURLOverride = baseURL[0]
	}

	err := HTTPSend(par)
	return t.Token, err
}
