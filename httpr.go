package httpr

import (
	"archive/tar"
	"compress/gzip"
	"compress/zlib"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func WrapWithCompressionReader(r io.Reader, compressionType string) (io.Reader, error) {
	switch compressionType {
	case CompressionTar:
		return tar.NewReader(r), nil
	case CompressionDeflate:
		return zlib.NewReader(r)
	case CompressionGzip:
		return gzip.NewReader(r)
	case CompressionNone:
		return r, nil
	default:
		return r, fmt.Errorf("unsupported response compression format '%s'", compressionType)
	}
}

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

func BuildRequest(url, method, compression, format string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return req, err
	}

	if compression != CompressionNone {
		req.Header.Set("Content-Encoding", compression)
	}

	contentTypeValue := GetContentTypeHeaderValue(format)
	if contentTypeValue != "" {
		req.Header.Set("Content-Type", contentTypeValue)
	}

	req.Header.Set("Accept", GetAcceptHeaderValue(compression, format))

	return req, nil
}

func DefaultClient() *http.Client {
	c := &http.Client{
		Timeout:   DefaultRequestTimeout,
		Transport: DefaultTransport(),
	}
	return c
}

func DoGetRequest(requestURL string) (Response, error) {
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	return DoRequest(context.Background(), DefaultClient(), req, CompressionNone)
}

func DoGetRequestWithRetry(requestURL string) (Response, error) {
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	return DoRequestWithRetry(context.Background(), DefaultClient(), req, CompressionNone)
}

func DoRequest(ctx context.Context, httpClient *http.Client, req *http.Request, compression string) (r Response, err error) {
	return doRequest(httpClient, req.WithContext(ctx))
}

func DoRequestWithRetry(ctx context.Context, httpClient *http.Client, req *http.Request, compression string) (Response, error) {
	req = req.WithContext(ctx)

	return NewHttpClient(
		*httpClient,
		WithRetryCount(DefaultRetriesCount),
		WithRetryDelay(DefaultRetryDelay),
	).Do(req)
}

func Is2xx(resp *http.Response) bool {
	return resp.StatusCode >= 200 && resp.StatusCode < 300
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
		delayMs := rand.Int63n(int64(delayLimit * 1000))
		time.Sleep(time.Millisecond * time.Duration(delayMs))
		return nil
	}
}
