package httpr

import (
	"net/http"
)

type basicAuthTransport struct {
	user      string
	password  string
	defaultTr http.Transport
}

func (tr *basicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header.Get("Authorization") == "" {
		req.SetBasicAuth(tr.user, tr.password)
	}

	return tr.defaultTr.RoundTrip(req)
}

func BuildBasicAuthTransport(username, password string) http.RoundTripper {
	return &basicAuthTransport{
		user:      username,
		password:  password,
		defaultTr: *DefaultTransport(),
	}
}

type bearerAuthTransport struct {
	token     string
	defaultTr http.Transport
}

func (tr *bearerAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+tr.token)
	return tr.defaultTr.RoundTrip(req)
}

func BuildBearerAuthTransport(token string) http.RoundTripper {
	return &bearerAuthTransport{
		token:     token,
		defaultTr: *DefaultTransport(),
	}
}

func DefaultTransport() *http.Transport {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.MaxConnsPerHost = _defaultConnsPerHost
	tr.MaxIdleConns = _defaultConnsPerHost
	tr.MaxIdleConnsPerHost = _defaultConnsPerHost
	tr.TLSHandshakeTimeout = _defaultTLSHandshakeTimeout

	return tr
}
