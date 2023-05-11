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
	DefaultRetriesCount         = 3
	DefaultRetryDelay           = time.Second * 3
	DefaultRetryDelayDelta      = 0
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
	preRequestHookFn PreRequestHookFn
}

func defaultClientSettings() *clientSettings {
	return &clientSettings{
		rateLimiter:      &unlimitedLimiter{},
		timeout:          DefaultRequestTimeout,
		retryCount:       0,
		retryDelay:       DefaultRetryDelay,
		retryDelayDelta:  DefaultRetryDelayDelta,
		redirectCheckFn:  nil,
		preRequestHookFn: func(req *http.Request) error { return nil },
	}
}

type Option func(settings *clientSettings)

var (
	_ Limiter  = (*unlimitedLimiter)(nil)
	_ Response = (*ClientResponse)(nil)
)

type RetryConditionFunc func(response *Response) bool

type Limiter interface {
	Take() time.Time
}

type unlimitedLimiter struct{}

func (l *unlimitedLimiter) Take() time.Time {
	return time.Now()
}

func WithRateLimiter(limiter Limiter) Option {
	return func(settings *clientSettings) {
		if limiter != nil {
			settings.rateLimiter = limiter
		}
	}
}

func WithRetryCount(retries int) Option {
	return func(settings *clientSettings) {
		settings.retryCount = retries
	}
}

func WithRetryDelay(delay time.Duration) Option {
	return func(settings *clientSettings) {
		settings.retryDelay = delay
	}
}

func WithRetryDelayDelta(delayDelta time.Duration) Option {
	return func(settings *clientSettings) {
		settings.retryDelayDelta = delayDelta
	}
}

func WithTransport(transport http.RoundTripper) Option {
	return func(settings *clientSettings) {
		if transport != nil {
			settings.transport = transport
		}
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(settings *clientSettings) {
		settings.timeout = timeout
	}
}

func WithCheckRedirect(checkFn func(*http.Request, []*http.Request) error) Option {
	return func(settings *clientSettings) {
		settings.redirectCheckFn = checkFn
	}
}

func WithRetryCondition(conditionFn RetryConditionFunc) Option {
	return func(settings *clientSettings) {
		settings.retryConditionFn = conditionFn
	}
}

func WithCookieJar(cookieJar http.CookieJar) Option {
	return func(settings *clientSettings) {
		settings.cookieJar = cookieJar
	}
}

type PreRequestHookFn func(req *http.Request) error

func WithPreRequestHook(hookFn PreRequestHookFn) Option {
	return func(settings *clientSettings) {
		settings.preRequestHookFn = hookFn
	}
}

type Client interface {
	Do(*http.Request) (Response, error)
	Get(string) (Response, error)
}

type client struct {
	client *http.Client

	PreRequestHookFn PreRequestHookFn
	RetryCount       int
	RetryDelay       time.Duration
	RetryDelayDelta  time.Duration
	RateLimiter      Limiter
	RequestTimeout   time.Duration
	RedirectCheckFn  func(*http.Request, []*http.Request) error
}

var _ Client = (*client)(nil)

func NewHttpClient(httpClient http.Client, opts ...Option) Client {
	settings := defaultClientSettings()
	for _, opt := range opts {
		opt(settings)
	}

	httpClient.Transport = settings.transport
	httpClient.Jar = settings.cookieJar

	return &client{
		client:           &httpClient,
		RetryCount:       settings.retryCount,
		RetryDelay:       settings.retryDelay,
		RateLimiter:      settings.rateLimiter,
		RequestTimeout:   settings.timeout,
		RedirectCheckFn:  settings.redirectCheckFn,
		PreRequestHookFn: settings.preRequestHookFn,
	}
}

func (c *client) Do(req *http.Request) (Response, error) {
	c.RateLimiter.Take()

	cancellableCtx, cancel := context.WithTimeout(req.Context(), c.RequestTimeout)
	defer cancel()

	if err := c.PreRequestHookFn(req); err != nil {
		return nil, err
	}

	if c.RetryCount <= 0 {
		return doRequest(c.client, req.WithContext(cancellableCtx))
	}

	var (
		retryErr  error
		retryTime = c.RetryDelay
		lastResp  Response
	)

	retryFunc := func() (Response, error) {
		for r := 0; r < c.RetryCount; r++ {
			resp, err := doRequest(c.client, req.WithContext(cancellableCtx))
			if err == nil && r <= c.RetryCount {
				return resp, err
			}
			if err != nil {
				retryErr = err
			}

			select {
			case <-time.After(c.RetryDelay):
				lastResp = resp
				retryTime += c.RetryDelayDelta
			case <-cancellableCtx.Done():
				return nil, cancellableCtx.Err()
			}
		}

		return lastResp, fmt.Errorf("failed to send request after %d attempt(s): %w", c.RetryCount, retryErr)
	}

	return retryFunc()
}

func (c *client) Get(requestURL string) (Response, error) {
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}

func (c *client) Post(requestURL, contentType string, body io.Reader) (Response, error) {
	req, err := http.NewRequest(http.MethodPost, requestURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Accept", contentType)
	return c.Do(req)
}

func (c *client) Client() *http.Client {
	return c.client
}

func (c *client) SetCookies(cookieOrigin *url.URL, cookies []*http.Cookie) {
	if c.client.Jar == nil {
		return
	}

	c.client.Jar.SetCookies(cookieOrigin, cookies)
}

func (c *client) SetTransport(transport http.RoundTripper) {
	c.client.Transport = transport
}

func doRequest(httpClient *http.Client, req *http.Request) (r *ClientResponse, err error) {
	r = new(ClientResponse)

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

	if Is4xx(r.rawResp.StatusCode) || Is5xx(r.rawResp.StatusCode) {
		return r, newResponseError("got http response error code", r.rawResp.StatusCode)
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
