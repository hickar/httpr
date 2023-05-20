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

	c := New()
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

func TestMethods(t *testing.T) {
	ts := createMethodTestServer()
	defer ts.Close()

	c := New()

	tests := []struct {
		name             string
		clientMethodCall func(*Client, string) (*Response, error)
		testURL          string
	}{
		{
			name: "TestGetMethod",
			clientMethodCall: func(c *Client, reqURL string) (*Response, error) {
				return c.Get(context.Background(), reqURL)
			},
			testURL: ts.URL + "/get",
		},
		{
			name: "TestPostMethod",
			clientMethodCall: func(c *Client, reqURL string) (*Response, error) {
				return c.Post(context.Background(), reqURL, nil)
			},
			testURL: ts.URL + "/post",
		},
		{
			name: "TestPutMethod",
			clientMethodCall: func(c *Client, reqURL string) (*Response, error) {
				return c.Put(context.Background(), reqURL, nil)
			},
			testURL: ts.URL + "/put",
		},
		{
			name: "TestPatchMethod",
			clientMethodCall: func(c *Client, reqURL string) (*Response, error) {
				return c.Patch(context.Background(), reqURL, nil)
			},
			testURL: ts.URL + "/patch",
		},
		{
			name: "TestDeleteMethod",
			clientMethodCall: func(c *Client, reqURL string) (*Response, error) {
				return c.Delete(context.Background(), reqURL)
			},
			testURL: ts.URL + "/delete",
		},
		{
			name: "TestOptionsMethod",
			clientMethodCall: func(c *Client, reqURL string) (*Response, error) {
				return c.Options(context.Background(), reqURL, nil)
			},
			testURL: ts.URL + "/options",
		},
		{
			name: "TestHeadMethod",
			clientMethodCall: func(c *Client, reqURL string) (*Response, error) {
				return c.Head(context.Background(), reqURL)
			},
			testURL: ts.URL + "/head",
		},
		{
			name: "TestTraceMethod",
			clientMethodCall: func(c *Client, reqURL string) (*Response, error) {
				return c.Head(context.Background(), reqURL)
			},
			testURL: ts.URL + "/trace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.clientMethodCall(&c, tt.testURL)
			if err != nil {
				t.Fatalf("expected nil error, got instead %v", err)
			}

			respCode := resp.StatusCode()
			if respCode != http.StatusOK {
				t.Fatalf("expected status code %d, got %d instead", http.StatusOK, respCode)
			}
		})
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
		WithRetryCount(expectedRetryCount),
		WithRetryDelay(0),
		WithRetryCondition(func(response *Response, err error) bool {
			return response.StatusCode() == http.StatusInternalServerError
		}),
	)
	resp, err := c.Get(context.Background(), ts.URL+"/test")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp.StatusCode() != http.StatusInternalServerError {
		t.Fatalf("expected status code %d, got %d instead", http.StatusInternalServerError, resp.StatusCode())
	}
	if actualRetryCount != expectedRetryCount {
		t.Fatalf("expected != actual, %d != %d", actualRetryCount, expectedRetryCount)
	}
}

func TestRequestTimeout(t *testing.T) {
	ts := createTestServer()
	defer ts.Close()

	c := New(WithTimeout(time.Second))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := c.Get(ctx, ts.URL+"/timeout")
	ts.CloseClientConnections()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context.Canceled error, but got error: %v", err)
	}
}

func TestGzipAutoUncompression(t *testing.T) {
	ts := createTestServer()
	defer ts.Close()

	c := New()

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/gzip-compressed", nil)
	req.Header.Set("Accept", AcceptGzipHeader)

	resp, err := c.Do(req, WithAutoDecompression(true))
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
