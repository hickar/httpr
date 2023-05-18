package httpr

import (
	"archive/tar"
	"compress/flate"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultRetryDelay           = time.Second * 3
	DefaultRequestTimeout       = time.Minute
	_defaultTLSHandshakeTimeout = time.Minute
	_defaultConnsPerHost        = 100
)

const (
	AcceptGzipHeader    = "application/gzip"
	AcceptTarHeader     = "application/x-tar"
	AcceptDeflateHeader = "application/zlib"
)

type clientSettings struct {
	rateLimiter      Limiter
	retryCount       int
	retryDelay       time.Duration
	retryDelayDelta  time.Duration
	retryConditionFn RetryConditionFunc
	timeout          time.Duration
	transport        http.RoundTripper
	cookieJar        http.CookieJar
	redirectCheckFn  func(*http.Request, []*http.Request) error

	preRequestHookFn  PreRequestHookFn
	postRequestHookFn PostRequestHookFn
}

var defaultClient = New(*http.DefaultClient)

func defaultClientSettings() *clientSettings {
	return &clientSettings{
		rateLimiter:       &unlimitedLimiter{},
		retryDelay:        DefaultRetryDelay,
		redirectCheckFn:   func(_ *http.Request, _ []*http.Request) error { return nil },
		preRequestHookFn:  func(_ *http.Request) error { return nil },
		postRequestHookFn: func(_ *http.Request, _ *Response) error { return nil },
	}
}

type Client struct {
	client *http.Client

	RetryCount      int
	RetryDelay      time.Duration
	RetryDelayDelta time.Duration
	RateLimiter     Limiter
	RedirectCheckFn func(*http.Request, []*http.Request) error

	PreRequestHookFn  PreRequestHookFn
	PostRequestHookFn PostRequestHookFn
}

func New(httpClient http.Client, opts ...Option) Client {
	settings := defaultClientSettings()
	for _, opt := range opts {
		opt(settings)
	}

	httpClient.Transport = settings.transport
	httpClient.Jar = settings.cookieJar

	return Client{
		client:            &httpClient,
		RetryCount:        settings.retryCount,
		RetryDelay:        settings.retryDelay,
		RateLimiter:       settings.rateLimiter,
		RedirectCheckFn:   settings.redirectCheckFn,
		PreRequestHookFn:  settings.preRequestHookFn,
		PostRequestHookFn: settings.postRequestHookFn,
	}
}

func (c *Client) Do(req *http.Request) (*Response, error) {
	c.RateLimiter.Take()

	if err := c.PreRequestHookFn(req); err != nil {
		return nil, err
	}

	var (
		retryErr   error
		retryTime  = c.RetryDelay
		retryCount = c.RetryCount
		ctx        = req.Context()
	)

	if retryCount < 1 {
		retryCount = 1
	}

	for r := 0; r < retryCount; r++ {
		resp, err := doRequest(c.client, req)
		if err == nil && r <= c.RetryCount {
			return resp, err
		}
		if err != nil {
			retryErr = err
		}

		select {
		case <-time.After(c.RetryDelay):
			retryTime += c.RetryDelayDelta
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return nil, fmt.Errorf("failed to send request after %d attempt(s): %w", c.RetryCount, retryErr)
}

func (c *Client) Get(ctx context.Context, requestURL string) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}

func (c *Client) Post(ctx context.Context, requestURL, contentType string, body io.Reader) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Accept", contentType)
	return c.Do(req)
}

func (c *Client) Put(ctx context.Context, requestURL string, body io.Reader) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, body)
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}

func (c *Client) Head(ctx context.Context, requestURL string) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, requestURL, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}

func (c *Client) Options(ctx context.Context, requestURL string) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodOptions, requestURL, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}

func (c *Client) Delete(ctx context.Context, requestURL string) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, requestURL, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}

func (c *Client) Client() *http.Client {
	return c.client
}

func (c *Client) SetCookies(cookieOrigin *url.URL, cookies []*http.Cookie) {
	if c.client.Jar == nil {
		return
	}

	c.client.Jar.SetCookies(cookieOrigin, cookies)
}

func (c *Client) SetTransport(transport http.RoundTripper) {
	c.client.Transport = transport
}

func doRequest(httpClient *http.Client, req *http.Request) (r *Response, err error) {
	r = new(Response)

	r.rawResp, err = httpClient.Do(req)
	if err != nil {
		return r, err
	}
	defer func(body io.Closer) {
		closeErr := body.Close()
		if closeErr != nil {
			err = closeErr
		}
	}(r.rawResp.Body)

	reader, err := wrapWithCompressionReader(r.rawResp, req)
	if err != nil {
		return r, fmt.Errorf("unable to wrap response in compression reader: %w", err)
	}
	closer, ok := reader.(io.Closer)
	if ok {
		defer func(body io.Closer) {
			closeErr := body.Close()
			if closeErr != nil {
				err = closeErr
			}
		}(closer)
	}

	r.body, err = io.ReadAll(reader)
	if err != nil {
		return r, fmt.Errorf("failed to read response bytes: %w", err)
	}

	return r, nil
}

func wrapWithCompressionReader(resp *http.Response, req *http.Request) (io.Reader, error) {
	for _, mimeType := range req.Header.Values("Accept") {
		switch strings.ToLower(mimeType) {
		case AcceptGzipHeader:
			return gzip.NewReader(resp.Body)

		case AcceptDeflateHeader:
			return flate.NewReader(resp.Body), nil

		case AcceptTarHeader:
			return tar.NewReader(resp.Body), nil

		}
	}

	for _, encodingType := range resp.Header.Values("Content-Encoding") {
		switch strings.ToLower(encodingType) {
		case CompressionGzip:
			return gzip.NewReader(resp.Body)

		case CompressionDeflate:
			return flate.NewReader(resp.Body), nil

		case CompressionTar:
			return tar.NewReader(resp.Body), nil

		}
	}

	return resp.Body, nil
}

type RetryConditionFunc func(resp *Response) bool

type Limiter interface {
	Take() time.Time
}
