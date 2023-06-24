package httpr

import (
	"net/http"
	"time"
)

const (
	_defaultTLSHandshakeTimeout = time.Minute
	_defaultConnsPerHost        = 100
)

var DefaultClient = New()

func New(opts ...Option) Client {
	return NewWithClient(&http.Client{}, opts...)
}

func NewWithClient(httpClient *http.Client, opts ...Option) Client {
	settings := newDefaultSettings()
	for _, opt := range opts {
		opt(&settings)
	}

	if httpClient == nil {
		httpClient = &http.Client{}
	}

	httpClient.Transport = settings.transport
	httpClient.Jar = settings.cookieJar

	return Client{
		client:   httpClient,
		settings: settings,
	}
}
