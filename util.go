package httpr

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func wrapWithCompressionReader(resp *http.Response, req *http.Request) (io.ReadCloser, error) {
	for _, mimeType := range req.Header.Values("Accept") {
		if strings.ToLower(mimeType) == "application/gzip" {
			return gzip.NewReader(resp.Body)
		}
	}

	return resp.Body, nil
}

func buildRequest(ctx context.Context, requestURL, method string, body any) (*http.Request, error) {
	reqBody, err := convertBodyToReader(body)
	if err != nil {
		return nil, fmt.Errorf("failed to build request body: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, reqBody)
	if err != nil {
		return req, err
	}

	return req, nil
}

func convertBodyToReader(body any) (io.Reader, error) {
	var reqBody io.Reader
	switch b := body.(type) {
	case string:
		reqBody = strings.NewReader(b)
	case []byte:
		reqBody = bytes.NewReader(b)
	case map[string]any:
		reqBodyBytes, err := json.Marshal(&b)
		if err != nil {
			return nil, err
		}

		reqBody = bytes.NewReader(reqBodyBytes)
	case io.Reader:
		reqBody = b
	}

	return reqBody, nil
}

// Do executes provided request by using DefaultClient.
func Do(req *http.Request, opts ...Option) (*Response, error) {
	return DefaultClient.Do(req, opts...)
}

// Is1xx check whether provided status code is in range of 100 and 200.
func Is1xx(code int) bool {
	return code >= 100 && code < 200
}

// Is2xx check whether provided status code is in range of 100 and 200.
func Is2xx(code int) bool {
	return code >= 200 && code < 300
}

// Is3xx check whether provided status code is in range of 100 and 200.
func Is3xx(code int) bool {
	return code >= 300 && code < 400
}

// Is4xx check whether provided status code is in range of 100 and 200.
func Is4xx(code int) bool {
	return code >= 400 && code < 500
}

// Is5xx check whether provided status code is in range of 100 and 200.
func Is5xx(code int) bool {
	return code >= 500 && code < 600
}

// IsValidURL checks whether provided URL is valid or not.
func IsValidURL(rawURL string) bool {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}

	if strings.TrimSpace(parsedURL.Host) == "" {
		return false
	}

	if len(strings.Split(parsedURL.Host, ".")) < 2 {
		return false
	}

	if strings.Contains(parsedURL.Host, "http:") || strings.Contains(parsedURL.Host, "https:") {
		return false
	}

	return true
}

// RandomDelay is pre-built PreRequestHookFn compliant function used for delaying request execution
// by random delay. Random delay is calculated by calling rand.Int63n with provided delayLimit argument value.
func RandomDelay(delayLimit time.Duration) PreRequestHookFn {
	return func(req *http.Request) error {
		if delayLimit < 0 {
			return nil
		}

		//nolint:gosec
		delayMs := rand.Int63n(int64(delayLimit))
		time.Sleep(time.Millisecond * time.Duration(delayMs))
		return nil
	}
}
