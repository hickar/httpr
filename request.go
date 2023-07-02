package httpr

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// RequestBuilder struct provides convenient interface
// for *http.Request instances construction.
type RequestBuilder struct {
	err error

	ctx                  context.Context
	url                  *url.URL
	method               string
	body                 any
	headers              map[string][]string
	queryParams          url.Values
	cookies              []*http.Cookie
	basicAuthCredentials *struct {
		user string
		pass string
	}
}

// NewRequest creates new RequestBuilder instance, which used for
// http.Request building.
func NewRequest() *RequestBuilder {
	return &RequestBuilder{
		headers:     make(map[string][]string),
		queryParams: make(url.Values),
	}
}

// SetURL sets target URL for current request.
func (rb *RequestBuilder) SetURL(requestURL string) *RequestBuilder {
	rb.url, rb.err = parseURL(requestURL)
	return rb
}

// SetMethod sets method for current request.
func (rb *RequestBuilder) SetMethod(method string) *RequestBuilder {
	rb.method = method
	return rb
}

// Get method create "get" HTTP request with specified body.
func (rb *RequestBuilder) Get(requestURL string, body any) *RequestBuilder {
	rb.method = http.MethodGet
	rb.SetURL(requestURL)
	rb.body = body
	return rb
}

// Post method create "post" HTTP request with specified body.
func (rb *RequestBuilder) Post(requestURL string, body any) *RequestBuilder {
	rb.method = http.MethodPost
	rb.SetURL(requestURL)
	rb.body = body
	return rb
}

// Patch method create "patch" HTTP request with specified body.
func (rb *RequestBuilder) Patch(requestURL string, body any) *RequestBuilder {
	rb.method = http.MethodPatch
	rb.SetURL(requestURL)
	rb.body = body
	return rb
}

// Put method create "put" HTTP request with specified body.
func (rb *RequestBuilder) Put(requestURL string, body any) *RequestBuilder {
	rb.method = http.MethodPut
	rb.SetURL(requestURL)
	rb.body = body
	return rb
}

// Options method create "options" HTTP request with specified body.
func (rb *RequestBuilder) Options(requestURL string, body any) *RequestBuilder {
	rb.method = http.MethodOptions
	rb.SetURL(requestURL)
	rb.body = body
	return rb
}

// Head method create "head" HTTP request.
func (rb *RequestBuilder) Head(requestURL string) *RequestBuilder {
	rb.method = http.MethodHead
	rb.SetURL(requestURL)
	return rb
}

// Connect method create "connect" HTTP request.
func (rb *RequestBuilder) Connect(requestURL string) *RequestBuilder {
	rb.method = http.MethodConnect
	rb.SetURL(requestURL)
	return rb
}

// Delete method create "delete" HTTP request with specified body.
func (rb *RequestBuilder) Delete(requestURL string, body any) *RequestBuilder {
	rb.method = http.MethodDelete
	rb.SetURL(requestURL)
	rb.body = body
	return rb
}

// Trace method create "trace" HTTP request.
func (rb *RequestBuilder) Trace(requestURL string) *RequestBuilder {
	rb.method = http.MethodTrace
	rb.SetURL(requestURL)
	return rb
}

// SetBody method sets body for current request.
// Body can be one of following concrete types or types, which implement
// interfaces: string, []byte, io.Reader.
func (rb *RequestBuilder) SetBody(body any) *RequestBuilder {
	rb.body = body
	return rb
}

// SetContext sets context for current request. If provided context is nil,
// new one will be created with context.Background().
func (rb *RequestBuilder) SetContext(ctx context.Context) *RequestBuilder {
	rb.ctx = ctx
	return rb
}

// SetHeader sets header with provided key and value.
func (rb *RequestBuilder) SetHeader(key, value string) *RequestBuilder {
	if rb.headers == nil {
		rb.headers = make(map[string][]string)
	}

	rb.headers[key] = append(rb.headers[key], value)
	return rb
}

// SetHeaders creates and sets headers for each key/value pair in provided map.
func (rb *RequestBuilder) SetHeaders(headers map[string]string) *RequestBuilder {
	for key, value := range headers {
		rb.SetHeader(key, value)
	}

	return rb
}

// SetQueryString provides option to set query string parameters by passing
// raw string.
func (rb *RequestBuilder) SetQueryString(query string) *RequestBuilder {
	if rb.queryParams == nil {
		rb.queryParams = make(url.Values)
	}

	queryParams, err := url.ParseQuery(query)
	if err != nil {
		rb.err = fmt.Errorf("malformed query: %w", err)
		return rb
	}

	for key, values := range queryParams {
		for _, value := range values {
			rb.queryParams.Add(key, value)
		}
	}

	return rb
}

// SetQueryParam sets query parameter with following key and value.
func (rb *RequestBuilder) SetQueryParam(key, value string) *RequestBuilder {
	if strings.TrimSpace(key) == "" {
		return rb
	}

	if rb.queryParams == nil {
		rb.queryParams = make(url.Values)
	}

	rb.queryParams.Set(key, value)
	return rb
}

// SetQueryParams sets multiple query parameters by calling SetQueryParams for each
// key/value in map.
func (rb *RequestBuilder) SetQueryParams(params map[string]string) *RequestBuilder {
	if rb.queryParams == nil {
		rb.queryParams = make(url.Values, len(params))
	}

	for key, value := range params {
		rb.SetQueryParam(key, value)
	}

	return rb
}

// SetCookies sets cookies for current request.
func (rb *RequestBuilder) SetCookies(cookies []*http.Cookie) *RequestBuilder {
	rb.cookies = cookies
	return rb
}

// SetBasicAuth encodes and sets basic HTTP authentication credentials.
func (rb *RequestBuilder) SetBasicAuth(user, pass string) *RequestBuilder {
	rb.basicAuthCredentials = &struct {
		user string
		pass string
	}{
		user: user,
		pass: pass,
	}
	return rb
}

// Build composes *http.Request instance. If errors occurred during previous building steps,
// they will be returned.
func (rb *RequestBuilder) Build() (*http.Request, error) {
	if rb.err != nil {
		return nil, rb.err
	}
	if rb.url == nil {
		return nil, errors.New("request url is not set")
	}

	reqURL := composeURL(rb.url, rb.queryParams)
	reqBody := convertBodyToReader(rb.body)
	reqMethod := composeMethod(rb.method)

	reqCtx := rb.ctx
	if reqCtx == nil {
		reqCtx = context.Background()
	}

	req, err := http.NewRequestWithContext(reqCtx, reqMethod, reqURL, reqBody)
	if err != nil {
		return nil, err
	}

	if rb.basicAuthCredentials != nil {
		req.SetBasicAuth(rb.basicAuthCredentials.user, rb.basicAuthCredentials.pass)
	}

	for key, values := range rb.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	for _, cookie := range rb.cookies {
		req.AddCookie(cookie)
	}

	return req, nil
}

func composeURL(reqURL *url.URL, params url.Values) string {
	encodedQuery := params.Encode()
	if encodedQuery == "" {
		return reqURL.String()
	}

	if reqURL.RawQuery == "" {
		reqURL.RawQuery = encodedQuery
	} else {
		reqURL.RawQuery += "&" + encodedQuery
	}

	return reqURL.String()
}

func composeMethod(method string) string {
	if method == "" {
		return http.MethodGet
	}

	return strings.ToUpper(method)
}

func parseURL(requestURL string) (*url.URL, error) {
	if !IsValidURL(requestURL) {
		return nil, fmt.Errorf("invalid URL '%s'", requestURL)
	}

	return url.Parse(requestURL)
}
