package httpr

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	AcceptHeader      = "Accept"
	ContentTypeHeader = "Content-Type"
)

type RequestBuilder struct {
	err error

	ctx         context.Context
	url         *url.URL
	method      string
	body        any
	headers     map[string][]string
	queryParams url.Values
}

func NewRequest() *RequestBuilder {
	return &RequestBuilder{
		headers:     make(map[string][]string),
		queryParams: make(url.Values),
	}
}

func (rb *RequestBuilder) SetURL(requestURL string) *RequestBuilder {
	rb.url, rb.err = parseURL(requestURL)
	return rb
}

func (rb *RequestBuilder) SetMethod(method string) *RequestBuilder {
	rb.method = method
	return rb
}

func (rb *RequestBuilder) Get(requestURL string, body any) *RequestBuilder {
	rb.method = http.MethodGet
	rb.SetURL(requestURL)
	rb.body = body
	return rb
}

func (rb *RequestBuilder) Post(requestURL string, body any) *RequestBuilder {
	rb.method = http.MethodPost
	rb.SetURL(requestURL)
	rb.body = body
	return rb
}

func (rb *RequestBuilder) Patch(requestURL string, body any) *RequestBuilder {
	rb.method = http.MethodPatch
	rb.SetURL(requestURL)
	rb.body = body
	return rb
}

func (rb *RequestBuilder) Put(requestURL string, body any) *RequestBuilder {
	rb.method = http.MethodPut
	rb.SetURL(requestURL)
	rb.body = body
	return rb
}

func (rb *RequestBuilder) Options(requestURL string, body any) *RequestBuilder {
	rb.method = http.MethodOptions
	rb.SetURL(requestURL)
	rb.body = body
	return rb
}

func (rb *RequestBuilder) Head(requestURL string) *RequestBuilder {
	rb.method = http.MethodHead
	rb.SetURL(requestURL)
	return rb
}

func (rb *RequestBuilder) Connect(requestURL string) *RequestBuilder {
	rb.method = http.MethodConnect
	rb.SetURL(requestURL)
	return rb
}

func (rb *RequestBuilder) Delete(requestURL string, body any) *RequestBuilder {
	rb.method = http.MethodDelete
	rb.SetURL(requestURL)
	rb.body = body
	return rb
}

func (rb *RequestBuilder) Trace(requestURL string) *RequestBuilder {
	rb.method = http.MethodTrace
	rb.SetURL(requestURL)
	return rb
}

func (rb *RequestBuilder) SetBody(body any) *RequestBuilder {
	rb.body = body
	return rb
}

func (rb *RequestBuilder) SetContext(ctx context.Context) *RequestBuilder {
	rb.ctx = ctx
	return rb
}

func (rb *RequestBuilder) SetHeader(key, value string) *RequestBuilder {
	if rb.headers == nil {
		rb.headers = make(map[string][]string)
	}

	rb.headers[key] = append(rb.headers[key], value)
	return rb
}

func (rb *RequestBuilder) SetHeaders(headers map[string]string) *RequestBuilder {
	for key, value := range headers {
		rb.SetHeader(key, value)
	}

	return rb
}

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

func (rb *RequestBuilder) SetQueryParams(params map[string]string) *RequestBuilder {
	if rb.queryParams == nil {
		rb.queryParams = make(url.Values, len(params))
	}

	for key, value := range params {
		rb.SetQueryParam(key, value)
	}

	return rb
}

func (rb *RequestBuilder) SetAcceptType(typeName string) *RequestBuilder {
	rb.SetHeader(AcceptHeader, getMimeType(typeName))
	return rb
}

func (rb *RequestBuilder) SetContentType(typeName string) *RequestBuilder {
	rb.SetHeader(ContentTypeHeader, getMimeType(typeName))
	return rb
}

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

	for key, values := range rb.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
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

func getMimeType(name string) string {
	switch strings.ToLower(name) {
	case CompressionTar, CompressionGzip:
		return "application/gzip"
	case CompressionDeflate:
		return "application/zlib"
	case FormatCsv:
		return "text/csv"
	case FormatJson:
		return "application/json"
	case FormatXml:
		return "application/xml"
	default:
		return "*/*"
	}
}

func parseURL(requestURL string) (*url.URL, error) {
	if !IsValidURL(requestURL) {
		return nil, fmt.Errorf("invalid URL '%s'", requestURL)
	}

	return url.Parse(requestURL)
}
