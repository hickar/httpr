package httpr

import (
	"bytes"
	"io"
	"net/http"
)

// Response is a wrapper above standard http.Response objects, with some
// convenience methods.
type Response struct {
	rawResp *http.Response
	body    []byte
}

// Bytes returns byte slice representation of response body.
func (r *Response) Bytes() []byte {
	if r == nil || r.rawResp == nil || r.body == nil {
		return []byte{}
	}

	return r.body
}

// Reader returns io.Reader.
func (r *Response) Reader() io.Reader {
	if r.rawResp == nil {
		return bytes.NewReader([]byte{})
	}

	return bytes.NewReader(r.body)
}

// String returns string representation of response body. If underlying response is nil,
// returns an empty string.
func (r *Response) String() string {
	return string(r.Bytes())
}

// StatusCode returns HTTP status code of underlying response. If response is nil,
// returns 0.
func (r *Response) StatusCode() int {
	if r == nil || r.rawResp == nil {
		return 0
	}

	return r.rawResp.StatusCode
}

// Headers returns a map of headers.
func (r *Response) Headers() map[string]string {
	headers := make(map[string]string)
	if r.rawResp == nil {
		return headers
	}

	for key, values := range r.rawResp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	return headers
}

// Cookies returns slice of response cookies.
func (r *Response) Cookies() []*http.Cookie {
	if r.rawResp == nil {
		return nil
	}

	return r.rawResp.Cookies()
}

// RequestURL returns request original URL.
func (r *Response) RequestURL() string {
	if r == nil || r.rawResp == nil {
		return ""
	}

	return r.rawResp.Request.URL.String()
}

// Raw returns reference to underlying http.Response object. Call to this method handles control
// over original object to the caller.
func (r *Response) Raw() *http.Response {
	if r == nil {
		return nil
	}

	return r.rawResp
}
