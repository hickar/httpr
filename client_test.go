package httpr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	_testMsg = "test message error"
)

func TestSuccessfulResponse(t *testing.T) {
	ts := createTestServer()
	defer ts.Close()

	c := New(http.Client{})
	resp, err := c.Get(context.Background(), ts.URL+"/test")
	if err != nil {
		t.Fatalf("expected no error, but got error '%v'", err)
	}

	var testResp struct {
		Msg string `json:"msg"`
	}
	if err = json.Unmarshal(resp.Bytes(), &testResp); err != nil {
		t.Fatalf("unexpected error during response unmarshaling: %v", err)
	}

	if testResp.Msg != _testMsg {
		t.Fatalf("response string was malformed: expected '%s' but got '%s'", _testMsg, testResp.Msg)
	}

	if resp.StatusCode() != http.StatusOK {
		t.Fatalf("expected status code %d, but got %d", http.StatusOK, resp.StatusCode())
	}
}

func TestRequestRetry(t *testing.T) {
	var (
		expectedRetryCount = 3
		actualRetryCount   int
	)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/test" || req.Method != http.MethodGet {
			return
		}

		actualRetryCount++
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"msg": "%s"}`, _testMsg)))
	}))
	defer ts.Close()

	c := New(
		http.Client{},
		WithRetryCount(expectedRetryCount),
		WithRetryDelay(0),
	)
	_, err := c.Get(context.Background(), ts.URL+"/test")
	if err == nil {
		t.Errorf("expected non-nil error, got %v", err)
	}

	if actualRetryCount != expectedRetryCount {
		t.Errorf("expected != actual, %d != %d", actualRetryCount, expectedRetryCount)
	}
}

func TestRequestTimeout(t *testing.T) {
	ts := createTestServer()
	defer ts.Close()

	c := New(http.Client{}, WithTimeout(time.Second))

	_, err := c.Get(context.Background(), ts.URL+"/timeout")
	ts.CloseClientConnections()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context.Canceled error, but got error: %v", err)
	}
}

func TestResponseDataUncompression(t *testing.T) {
	ts := createTestServer()
	defer ts.Close()

	c := New(http.Client{})

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/gzip-compressed", nil)
	req.Header.Set("Accept", AcceptGzipHeader)

	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("expected no error but got error '%+v'", err)
	}

	var testResp struct {
		Msg string `json:"msg"`
	}
	if err = json.Unmarshal(resp.Bytes(), &testResp); err != nil {
		t.Fatalf("failed to unmarshal response body with following error: %v", err)
	}

	if testResp.Msg != _testMsg {
		t.Fatalf("expected message '%s', but got '%s'", _testMsg, testResp.Msg)
	}
}
