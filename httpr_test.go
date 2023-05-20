package httpr

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	getPayloadSuccess = "sample"
)

func handleTest(w http.ResponseWriter) {
	_, _ = w.Write([]byte(fmt.Sprintf(`{"msg": "%s"}`, _testMsg)))
	w.WriteHeader(http.StatusOK)
}

func handleTestTimeout(w http.ResponseWriter, req *http.Request) {
	select {
	case <-time.After(time.Second * 10):
	case <-req.Context().Done():
	}

	w.WriteHeader(http.StatusRequestTimeout)
}

func handleGzipCompression(w io.Writer) {
	var (
		respData = []byte(fmt.Sprintf(`{"msg": "%s"}`, _testMsg))
		wr       = gzip.NewWriter(w)
	)

	if _, err := wr.Write(respData); err != nil {
		panic(err)
	}
	if err := wr.Close(); err != nil {
		panic(err)
	}
	if err := wr.Flush(); err != nil {
		panic(err)
	}
}

func handleMethodGet(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/get" || req.Method != http.MethodGet {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(getPayloadSuccess))
}

func createTestServer() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			switch req.URL.Path {
			case "/test":
				handleTest(w)
			case "/timeout":
				handleTestTimeout(w, req)
			case "/gzip-compressed":
				handleGzipCompression(w)
			case "/get":
				handleMethodGet(w, req)
			}
		}

		w.WriteHeader(http.StatusInternalServerError)
	}))

	return ts
}

func createMethodTestServer() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/get":
			if req.Method == http.MethodGet {
				w.WriteHeader(http.StatusOK)
				return
			}
		case "/post":
			if req.Method == http.MethodPost {
				w.WriteHeader(http.StatusOK)
				return
			}
		case "/patch":
			if req.Method == http.MethodPatch {
				w.WriteHeader(http.StatusOK)
				return
			}
		case "/put":
			if req.Method == http.MethodPut {
				w.WriteHeader(http.StatusOK)
				return
			}
		case "/delete":
			if req.Method == http.MethodDelete {
				w.WriteHeader(http.StatusOK)
				return
			}
		case "/head":
			if req.Method == http.MethodHead {
				w.WriteHeader(http.StatusOK)
				return
			}
		case "/options":
			if req.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
		case "/connect":
			if req.Method == http.MethodConnect {
				w.WriteHeader(http.StatusOK)
				return
			}
		case "/trace":
			if req.Method == http.MethodTrace {
				w.WriteHeader(http.StatusOK)
				return
			}
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusInternalServerError)
	}))

	return ts
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid_HTTPS",
			input:    "https://test.com",
			expected: true,
		},
		{
			name:     "Valid_HTTP",
			input:    "http://test.com",
			expected: true,
		},
		{
			name:     "Valid_WithQueryParams",
			input:    "https://test.com?param1=1&param2=2",
			expected: true,
		},
		{
			name:     "Invalid_DuplicateQueryStartFlag",
			input:    "https://text.com???",
			expected: true,
		},
		{
			name:     "Invalid_NoScheme",
			input:    "test.com",
			expected: false,
		},
		{
			name:     "Invalid_NoDomain",
			input:    ".test",
			expected: false,
		},
		{
			name:     "Invalid_SchemeOnly",
			input:    "https://",
			expected: false,
		},
		{
			name:     "Invalid_NotCompletedDomain",
			input:    "https://test",
			expected: false,
		},
		{
			name:     "Invalid_DuplicateScheme",
			input:    "https://https://test.com",
			expected: false,
		},
		{
			name:     "Invalid_DuplicateSchemeDifferentCase",
			input:    "https://HTTps://test.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := IsValidURL(tt.input)

			if tt.expected != actual {
				t.Fatalf("expected != actual: %t != %t", tt.expected, actual)
			}
		})
	}
}
