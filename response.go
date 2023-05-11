package httpr

import (
	"bytes"
	"io"
	"net/http"
)

type Response interface {
	Bytes() []byte
	Reader() io.Reader
	String() string
	StatusCode() int
	Headers() map[string]string
	Cookies() []*http.Cookie
	RequestURL() string
}

type ClientResponse struct {
	rawResp *http.Response
	body    []byte
}

func (r *ClientResponse) Bytes() []byte {
	if r == nil || r.rawResp == nil || r.body == nil {
		return []byte{}
	}

	return r.body
}

func (r *ClientResponse) Reader() io.Reader {
	if r.rawResp == nil {
		return bytes.NewReader([]byte{})
	}

	return bytes.NewReader(r.body)
}

func (r *ClientResponse) String() string {
	return string(r.Bytes())
}

func (r *ClientResponse) StatusCode() int {
	if r.rawResp == nil {
		return 0
	}

	return r.rawResp.StatusCode
}

func (r *ClientResponse) Headers() map[string]string {
	headers := make(map[string]string)
	if r.rawResp == nil {
		return headers
	}

	for key, values := range r.rawResp.Header {
		headers[key] = values[0]
	}
	return headers
}

func (r *ClientResponse) Cookies() []*http.Cookie {
	if r.rawResp == nil {
		return nil
	}

	return r.rawResp.Cookies()
}

func (r *ClientResponse) RequestURL() string {
	if r == nil || r.rawResp == nil {
		return ""
	}

	return r.rawResp.Request.URL.String()
}
