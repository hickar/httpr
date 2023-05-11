package httpr

import (
	"context"
	"io"
	"net/http"
	"strings"
)

const (
	AcceptHeader      = "Accept"
	ContentTypeHeader = "Content-Type"
)

type RequestBuilder struct {
	err error

	ctx     context.Context
	url     string
	method  string
	body    io.Reader
	headers map[string][]string
}

func NewRequestBuilder() RequestBuilder {
	return RequestBuilder{
		headers: make(map[string][]string),
	}
}

func (rb *RequestBuilder) SetUrl(requestUrl string) *RequestBuilder {
	rb.url = requestUrl
	return rb
}

func (rb *RequestBuilder) SetMethod(method string) *RequestBuilder {
	rb.method = method
	return rb
}

func (rb *RequestBuilder) Get(requestUrl string) *RequestBuilder {
	rb.method = http.MethodGet
	rb.url = requestUrl
	return rb
}

func (rb *RequestBuilder) Post(requestUrl string, body io.Reader) *RequestBuilder {
	rb.method = http.MethodPost
	rb.url = requestUrl
	rb.body = body
	return rb
}

func (rb *RequestBuilder) SetBody(body io.Reader) *RequestBuilder {
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

	req, err := http.NewRequest(rb.method, rb.url, rb.body)
	if err != nil {
		return nil, err
	}

	if rb.ctx != nil {
		req = req.WithContext(rb.ctx)
	}

	for key, values := range rb.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	return req, nil
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
