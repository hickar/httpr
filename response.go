package httpr

import (
	"bytes"
	"io"
	"net/http"
)

type Response struct {
	rawResp *http.Response
	body    []byte
}

func (r *Response) Bytes() []byte {
	if r == nil || r.rawResp == nil || r.body == nil {
		return []byte{}
	}

	return r.body
}

func (r *Response) Reader() io.Reader {
	if r.rawResp == nil {
		return bytes.NewReader([]byte{})
	}

	return bytes.NewReader(r.body)
}

func (r *Response) String() string {
	return string(r.Bytes())
}

func (r *Response) StatusCode() int {
	if r.rawResp == nil {
		return 0
	}

	return r.rawResp.StatusCode
}

func (r *Response) Headers() map[string]string {
	headers := make(map[string]string)
	if r.rawResp == nil {
		return headers
	}

	for key, values := range r.rawResp.Header {
		headers[key] = values[0]
	}
	return headers
}

func (r *Response) Cookies() []*http.Cookie {
	if r.rawResp == nil {
		return nil
	}

	return r.rawResp.Cookies()
}

func (r *Response) RequestURL() string {
	if r == nil || r.rawResp == nil {
		return ""
	}

	return r.rawResp.Request.URL.String()
}
