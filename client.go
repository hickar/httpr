package httpr

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	client   *http.Client
	settings clientSettings
}

type clientSettings struct {
	rateLimiter          Limiter
	retryCount           int
	retryDelay           time.Duration
	retryDelayDelta      time.Duration
	retryConditionFn     RetryConditionFunc
	timeout              time.Duration
	transport            http.RoundTripper
	cookieJar            http.CookieJar
	decompressionEnabled bool

	redirectCheckFn   func(*http.Request, []*http.Request) error
	preRequestHookFn  PreRequestHookFn
	postRequestHookFn PostRequestHookFn
}

func (c *Client) Do(req *http.Request, opts ...Option) (*Response, error) {
	settings := c.settings
	if len(opts) > 0 {
		settings = newDefaultSettings()
		for _, opt := range opts {
			opt(&settings)
		}
	}

	settings.rateLimiter.Take()

	if err := settings.preRequestHookFn(req); err != nil {
		return nil, err
	}

	var (
		ctx        = req.Context()
		resp       *Response
		err        error
		retryTime  = settings.retryDelay
		retryCount = settings.retryCount
	)

	if retryCount < 1 {
		retryCount = 1
	}

	for r := 0; r < retryCount; r++ {
		resp, err = doRequest(c.client, req, settings)
		settings.postRequestHookFn(req, resp)

		mustRetry := settings.retryConditionFn(resp, err)
		if !mustRetry {
			break
		}
		if err == nil && r <= retryCount && !mustRetry {
			return resp, err
		}

		select {
		case <-time.After(settings.retryDelay):
			retryTime += settings.retryDelayDelta
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to send request after %d attempt(s): %w", settings.retryCount, err)
	}

	return resp, nil
}

func (c *Client) Get(ctx context.Context, requestURL string, opts ...Option) (*Response, error) {
	req, err := buildRequest(ctx, requestURL, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req, opts...)
}

func (c *Client) Post(ctx context.Context, requestURL string, body any, opts ...Option) (*Response, error) {
	req, err := buildRequest(ctx, requestURL, http.MethodPost, body)
	if err != nil {
		return nil, err
	}

	return c.Do(req, opts...)
}

func (c *Client) Put(ctx context.Context, requestURL string, body any, opts ...Option) (*Response, error) {
	req, err := buildRequest(ctx, requestURL, http.MethodPut, body)
	if err != nil {
		return nil, err
	}

	return c.Do(req, opts...)
}

func (c *Client) Patch(ctx context.Context, requestURL string, body any, opts ...Option) (*Response, error) {
	req, err := buildRequest(ctx, requestURL, http.MethodPatch, body)
	if err != nil {
		return nil, err
	}

	return c.Do(req, opts...)
}

func (c *Client) Head(ctx context.Context, requestURL string, opts ...Option) (*Response, error) {
	req, err := buildRequest(ctx, requestURL, http.MethodHead, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req, opts...)
}

func (c *Client) Options(ctx context.Context, requestURL string, body any, opts ...Option) (*Response, error) {
	req, err := buildRequest(ctx, requestURL, http.MethodOptions, body)
	if err != nil {
		return nil, err
	}

	return c.Do(req, opts...)
}

func (c *Client) Connect(ctx context.Context, requestURL string, opts ...Option) (*Response, error) {
	req, err := buildRequest(ctx, requestURL, http.MethodConnect, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req, opts...)
}

func (c *Client) Delete(ctx context.Context, requestURL string, opts ...Option) (*Response, error) {
	req, err := buildRequest(ctx, requestURL, http.MethodDelete, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req, opts...)
}

func (c *Client) Trace(ctx context.Context, requestURL string, opts ...Option) (*Response, error) {
	req, err := buildRequest(ctx, requestURL, http.MethodTrace, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req, opts...)
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

func doRequest(httpClient *http.Client, req *http.Request, settings clientSettings) (*Response, error) {
	var (
		r   = new(Response)
		err error
	)

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

	reader := r.rawResp.Body
	if settings.decompressionEnabled {
		reader, err = wrapWithCompressionReader(r.rawResp, req)
		if err != nil {
			return r, fmt.Errorf("unable to wrap response in compression reader: %w", err)
		}
	}

	defer func(body io.Closer) {
		closeErr := body.Close()
		if closeErr != nil {
			err = closeErr
		}
	}(reader)

	r.body, err = io.ReadAll(reader)
	if err != nil {
		return r, fmt.Errorf("failed to read response bytes: %w", err)
	}

	return r, nil
}
