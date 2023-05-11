package httpr

import (
	"archive/tar"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func handleTest(w http.ResponseWriter) {
	_, _ = w.Write([]byte(fmt.Sprintf(`{"msg": "%s"}`, _testMsg)))
	w.WriteHeader(http.StatusOK)
}

func handleTestTimeout(w http.ResponseWriter, req *http.Request) {
	select {
	case <-time.After(time.Minute):
	case <-req.Context().Done():
	}

	w.WriteHeader(http.StatusRequestTimeout)
}

func handleTestCompressed(compression string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var (
			respData = []byte(fmt.Sprintf(`{"msg": "%s"}`, _testMsg))
			wr       io.WriteCloser
		)

		switch compression {
		case CompressionGzip:
			wr = gzip.NewWriter(w)
		case CompressionTar:
			wr = tar.NewWriter(w)
		case CompressionDeflate:
			wr = zlib.NewWriter(w)
		default:
			panic("invalid compression type: " + compression)
		}

		_, err := wr.Write(respData)
		if err != nil {
			panic(err)
		}
		defer func(wr io.WriteCloser) {
			if err = wr.Close(); err != nil {
				panic(err)
			}

			flusher, ok := wr.(interface{ Flush() error })
			if ok {
				return
			}

			if err = flusher.Flush(); err != nil {
				panic(err)
			}
		}(wr)

		flusher, ok := w.(http.Flusher)
		if ok {
			flusher.Flush()
		}
	}
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
				handleTestCompressed(CompressionGzip)(w, req)
			case "/tar-compressed":
				handleTestCompressed(CompressionTar)(w, req)
			case "/deflate-compressed":
				handleTestCompressed(CompressionDeflate)(w, req)
			}
		}
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
			assert.Equal(t, tt.expected, actual)
		})
	}
}
