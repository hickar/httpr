package httpr

import (
	"net/http"
)

type basicAuthTransport struct {
	user string
	pass string
	tr   http.RoundTripper
}

func (tr *basicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header.Get("Authorization") == "" {
		req.SetBasicAuth(tr.user, tr.pass)
	}

	return tr.tr.RoundTrip(req)
}

// NewBasicAuthTransport creates http.Transport wrapper, which adds basic authentication
// credentials to 'Authorization' header before request is being sent.
func NewBasicAuthTransport(transport http.RoundTripper, user, pass string) http.RoundTripper {
	return &basicAuthTransport{
		user: user,
		pass: pass,
		tr:   transport,
	}
}

type bearerAuthTransport struct {
	token string
	tr    http.RoundTripper
}

func (tr *bearerAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+tr.token)
	return tr.tr.RoundTrip(req)
}

// NewBearerAuthTransport creates http.Transport wrapper, which adds authentication
// token to 'Authorization' header before request is being sent.
func NewBearerAuthTransport(transport http.RoundTripper, token string) http.RoundTripper {
	return &bearerAuthTransport{
		tr:    transport,
		token: token,
	}
}

// DefaultTransport creates slightly modified version of http.DefaultTransport.
// Maximum connections per host is set to 100.
// Maximum idle connections is set to 100.
// Maximum idle connections per host is set to 100.
// TLS handshake timeout is set to time.Minute.
func DefaultTransport() *http.Transport {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.MaxConnsPerHost = _defaultConnsPerHost
	tr.MaxIdleConns = _defaultConnsPerHost
	tr.MaxIdleConnsPerHost = _defaultConnsPerHost
	tr.TLSHandshakeTimeout = _defaultTLSHandshakeTimeout

	return tr
}
