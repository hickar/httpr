package httpr

import (
	"bytes"
	"net/http"
	"testing"
)

func TestResponseBytes(t *testing.T) {
	t.Run("NilResponse", func(t *testing.T) {
		var resp Response
		assertNoPanic(t, func() { resp.Bytes() })
	})

	t.Run("EmptyResponse", func(t *testing.T) {
		resp := (*Response)(nil)
		assertNoPanic(t, func() { resp.Bytes() })
	})

	t.Run("NoBody", func(t *testing.T) {
		resp := Response{body: []byte("content")}
		assertNoPanic(t, func() { resp.Bytes() })
	})

	t.Run("BodyContent", func(t *testing.T) {
		var (
			expectedBody = []byte("content")
			actualBody   = (&Response{
				rawResp: &http.Response{},
				body:    expectedBody,
			}).Bytes()
		)

		if !bytes.Equal(expectedBody, actualBody) {
			t.Errorf("expected body != actual: '%s' != '%s'", string(expectedBody), string(actualBody))
		}
	})
}

func TestResponseString(t *testing.T) {
	expectedBodyContent := "content"

	resp := &Response{
		body:    []byte(expectedBodyContent),
		rawResp: &http.Response{},
	}
	actualBodyContent := resp.String()

	if expectedBodyContent != actualBodyContent {
		t.Errorf("expected response string != actual: '%s' != '%s'", expectedBodyContent, actualBodyContent)
	}
}

func TestResponseStatusCode(t *testing.T) {
	t.Run("NilResponse", func(t *testing.T) {
		var (
			resp               = (*Response)(nil)
			expectedStatusCode = 0
			actualStatusCode   int
		)

		assertNoPanic(t, func() { actualStatusCode = resp.StatusCode() })

		if expectedStatusCode != actualStatusCode {
			t.Errorf("expected nil response to return status code 0, not %d", actualStatusCode)
		}
	})

	t.Run("OKResponse", func(t *testing.T) {
		var (
			expectedStatusCode = http.StatusOK
			actualStatusCode   int
			resp               = Response{rawResp: &http.Response{StatusCode: expectedStatusCode}}
		)

		actualStatusCode = resp.StatusCode()

		if expectedStatusCode != actualStatusCode {
			t.Errorf("expected status code %d, got %d instead", expectedStatusCode, actualStatusCode)
		}
	})
}

//nolint:thelper
func assertNoPanic(t *testing.T, fn func()) {
	defer func() {
		if recover() != nil {
			t.Errorf("panic occurred during function execution")
		}
	}()

	fn()
}
