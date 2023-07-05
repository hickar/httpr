// Copyright (c) 2023 Egor Pershin (hickar@protonmail.ch), All rights reserved.
// httpr source code and usage is governed by a MIT style
// license that can be found in the LICENSE file.

// Package httpr provides convenient methods for building and executing HTTP requests
// in GO idiomatic way.
package httpr

import (
	"net/http"
	"time"
)

const (
	_defaultTLSHandshakeTimeout = time.Minute
	_defaultConnsPerHost        = 100
)

// DefaultClient is static client initialized with call to New.
var DefaultClient = New()

// New creates new client with provided Options. Options must implement Option interface.
// Call to New is similar to call NewWithClient(&http.Client{}, opts...}.
func New(opts ...Option) Client {
	return NewWithClient(&http.Client{}, opts...)
}

// NewWithClient creates new client, which uses passed http.Client instance and options.
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
