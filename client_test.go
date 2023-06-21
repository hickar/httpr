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
	resp, err := c.Get(context.Background(), ts.URL+"/test", nil)
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
		expectedBody     string
	}{
		{
			name: "TestGetMethod",
			clientMethodCall: func(c *Client, reqURL string) (*Response, error) {
				return c.Get(context.Background(), reqURL, nil)
			},
			testURL:      ts.URL + "/get",
			expectedBody: methodSuccessBody,
		},
		{
			name: "TestPostMethod",
			clientMethodCall: func(c *Client, reqURL string) (*Response, error) {
				return c.Post(context.Background(), reqURL, nil)
			},
			testURL:      ts.URL + "/post",
			expectedBody: methodSuccessBody,
		},
		{
			name: "TestPutMethod",
			clientMethodCall: func(c *Client, reqURL string) (*Response, error) {
				return c.Put(context.Background(), reqURL, nil)
			},
			testURL:      ts.URL + "/put",
			expectedBody: methodSuccessBody,
		},
		{
			name: "TestPatchMethod",
			clientMethodCall: func(c *Client, reqURL string) (*Response, error) {
				return c.Patch(context.Background(), reqURL, nil)
			},
			testURL:      ts.URL + "/patch",
			expectedBody: methodSuccessBody,
		},
		{
			name: "TestDeleteMethod",
			clientMethodCall: func(c *Client, reqURL string) (*Response, error) {
				return c.Delete(context.Background(), reqURL)
			},
			testURL:      ts.URL + "/delete",
			expectedBody: methodSuccessBody,
		},
		{
			name: "TestOptionsMethod",
			clientMethodCall: func(c *Client, reqURL string) (*Response, error) {
				return c.Options(context.Background(), reqURL, nil)
			},
			testURL:      ts.URL + "/options",
			expectedBody: methodSuccessBody,
		},
		{
			name: "TestHeadMethod",
			clientMethodCall: func(c *Client, reqURL string) (*Response, error) {
				return c.Head(context.Background(), reqURL)
			},
			testURL:      ts.URL + "/head",
			expectedBody: "",
		},
		{
			name: "TestTraceMethod",
			clientMethodCall: func(c *Client, reqURL string) (*Response, error) {
				return c.Trace(context.Background(), reqURL)
			},
			testURL:      ts.URL + "/trace",
			expectedBody: methodSuccessBody,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.clientMethodCall(&c, tt.testURL)
			if err != nil {
				t.Errorf("expected nil error, got instead %v", err)
			}

			respCode := resp.StatusCode()
			if respCode != http.StatusOK {
				t.Errorf("expected status code %d, got %d instead", http.StatusOK, respCode)
			}

			respBody := string(resp.Bytes())
			if respBody != tt.expectedBody {
				t.Errorf("expected body content %q, got %q", methodSuccessBody, respBody)
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
	resp, err := c.Get(context.Background(), ts.URL+"/test", nil)
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

	_, err := c.Get(ctx, ts.URL+"/timeout", nil)
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

func TestPreRequestHookOption(t *testing.T) {
	t.Run("PreRequest_Basic", func(t *testing.T) {
		var preRequestHookWasCalled bool

		ts := createTestServer()
		defer ts.Close()

		client := New(WithPreRequestHook(func(req *http.Request) error {
			preRequestHookWasCalled = true
			return nil
		}))

		_, err := client.Get(context.Background(), ts.URL+"/test", nil)
		if err != nil {
			t.Error("should not return err")
		}

		if !preRequestHookWasCalled {
			t.Error("pre-request hook must have been called")
		}
	})

	t.Run("PreRequest_ShouldAbortRequest", func(t *testing.T) {
		ts := createTestServer()
		defer ts.Close()

		abortErr := errors.New("abort error")
		client := New(WithPreRequestHook(func(req *http.Request) error {
			return abortErr
		}))

		resp, err := client.Get(context.Background(), ts.URL+"/test", nil)
		if !errors.Is(err, abortErr) {
			t.Errorf("returned error should be equal to abortErr, got %v instead", err)
		}
		if resp != nil {
			t.Error("response should be nil")
		}
	})
}

func TestPostRequestHookOption(t *testing.T) {
	var postRequestHookWasCalled bool

	ts := createTestServer()
	defer ts.Close()

	client := New(WithPostRequestHook(func(_ *http.Request, _ *Response) {
		postRequestHookWasCalled = true
	}))

	_, _ = client.Get(context.Background(), ts.URL+"/test", nil)

	if !postRequestHookWasCalled {
		t.Error("post hook must have been called")
	}
}
