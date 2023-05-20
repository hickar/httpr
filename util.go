package httpr

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func GetContentTypeHeaderValue(formatType string) string {
	switch formatType {
	case FormatJson:
		return "application/json"
	case FormatXml:
		return "application/xml"
	case FormatCsv:
		return "text/csv"
	default:
		return ""
	}
}

func GetAcceptHeaderValue(compressionType, formatType string) string {
	if compressionType != CompressionNone {
		switch compressionType {
		case CompressionTar, CompressionGzip:
			return "application/gzip"
		case CompressionDeflate:
			return "application/zlib"
		default:
			return "*/*"
		}
	}

	if formatType != "" {
		switch formatType {
		case FormatCsv:
			return "text/csv"
		case FormatJson:
			return "application/json"
		case FormatXml:
			return "application/xml"
		}
	}

	return "*/*"
}

func wrapWithCompressionReader(resp *http.Response, req *http.Request) (io.ReadCloser, error) {
	for _, mimeType := range req.Header.Values("Accept") {
		if strings.ToLower(mimeType) == AcceptGzipHeader {
			return gzip.NewReader(resp.Body)
		}
	}

	return resp.Body, nil
}

func buildRequest(ctx context.Context, requestURL, method string, body any) (*http.Request, error) {
	var reqBody io.Reader
	switch b := body.(type) {
	case string:
		reqBody = strings.NewReader(b)
	case []byte:
		reqBody = bytes.NewReader(b)
	case io.Reader:
		reqBody = b
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, reqBody)
	if err != nil {
		return req, err
	}

	return req, nil
}

func Do(req *http.Request, opts ...Option) (*Response, error) {
	return DefaultClient.Do(req, opts...)
}

func Is1xx(code int) bool {
	return code >= 200 && code < 300
}

func Is2xx(code int) bool {
	return code >= 200 && code < 300
}

func Is3xx(code int) bool {
	return code >= 300 && code < 400
}

func Is4xx(code int) bool {
	return code >= 400 && code < 500
}

func Is5xx(code int) bool {
	return code >= 500 && code < 600
}

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

func RandomDelay(delayLimit float64) PreRequestHookFn {
	return func(req *http.Request) error {
		//nolint:gosec
		delayMs := rand.Int63n(int64(delayLimit * 1000))
		time.Sleep(time.Millisecond * time.Duration(delayMs))
		return nil
	}
}
