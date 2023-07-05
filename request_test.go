package httpr

import (
	"net/http"
	"testing"
)

func TestBuilderSetURL(t *testing.T) {
	var (
		rb          RequestBuilder
		expectedURL = "https://test.url.com"
	)

	req, _ := rb.SetURL(expectedURL).Build()
	if req.URL.String() != expectedURL {
		t.Errorf("expected %q request url, got %q", expectedURL, req.URL.String())
	}
}

func TestBuilderSetMethod(t *testing.T) {
	var rb RequestBuilder

	req, err := rb.
		SetMethod(http.MethodGet).
		SetURL("https://test.url.com").
		Build()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if req.Method != http.MethodGet {
		t.Errorf("expected method %q, got %q instead", http.MethodGet, req.Method)
	}
}

func TestBuilderSetHeaders(t *testing.T) {
	var rb RequestBuilder

	headers := map[string]string{
		"Authorization":  "Basic xxx",
		"Content-Type":   "application/json",
		"Content-Length": "42",
	}

	req, err := rb.
		SetHeaders(headers).
		SetURL("https://test.url.com").
		Build()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	for key, value := range req.Header {
		initHeaderValue, ok := headers[key]
		if !ok {
			t.Errorf("expected header %q does not exist", key)
		}

		if len(value) == 0 {
			t.Errorf("expected header %q to have some values", key)
		}

		if initHeaderValue != value[0] {
			t.Errorf("expected header %q to have value %q, not %q", key, initHeaderValue, value[0])
		}
	}
}

func TestBuilderSetQueryParams(t *testing.T) {
	var (
		rb          RequestBuilder
		expectedURL = "https://test.url.com?param1=value1&param2="
	)

	params := map[string]string{
		"param1": "value1",
		"param2": "",
		"":       "value3",
	}

	req, err := rb.
		SetQueryParams(params).
		SetURL("https://test.url.com").
		Build()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if req.URL.String() != expectedURL {
		t.Errorf("expected url %q, got %q instead", expectedURL, req.URL.String())
	}
}

func TestBuilderSetQueryString(t *testing.T) {
	tests := []struct {
		name        string
		inputQuery  string
		expectedURL string
	}{
		{
			name:        "TestQueryString_OneValue",
			inputQuery:  "param1=value1",
			expectedURL: "https://test.url.com?param1=value1",
		},
		{
			name:        "TestQueryString_MultipleValues",
			inputQuery:  "param1=value1&param2=value2",
			expectedURL: "https://test.url.com?param1=value1&param2=value2",
		},
		{
			name:        "TestQueryString_KeyOnly",
			inputQuery:  "param1",
			expectedURL: "https://test.url.com?param1=",
		},
		{
			name:        "TestQueryString_EmptyValues",
			inputQuery:  "&&&",
			expectedURL: "https://test.url.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rb RequestBuilder

			req, err := rb.
				SetURL("https://test.url.com").
				SetQueryString(tt.inputQuery).
				Build()
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			actualURL := req.URL.String()
			if actualURL != tt.expectedURL {
				t.Errorf("expected %v, got %v", tt.expectedURL, actualURL)
			}
		})
	}
}

func TestBuilderErrorOutput(t *testing.T) {
	tests := []struct {
		name              string
		buildRequestFn    func() (*http.Request, error)
		shouldReturnError bool
	}{
		{
			name: "NoError",
			buildRequestFn: func() (*http.Request, error) {
				return NewRequest().Get("https://test.com", nil).Build()
			},
			shouldReturnError: false,
		},
		{
			name: "Error_SetURL",
			buildRequestFn: func() (*http.Request, error) {
				return NewRequest().SetURL("").Build()
			},
			shouldReturnError: true,
		},
		{
			name: "Error_SetQueryString",
			buildRequestFn: func() (*http.Request, error) {
				return NewRequest().Get("https://test.com", nil).SetQueryString("bar=1&amp;baz=2").Build()
			},
			shouldReturnError: true,
		},
		{
			name: "Error_NilURL",
			buildRequestFn: func() (*http.Request, error) {
				return NewRequest().Build()
			},
			shouldReturnError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := tt.buildRequestFn()
			if err != nil != tt.shouldReturnError {
				t.Errorf("err != nil == %t, shouldReturnError == %t", err != nil, tt.shouldReturnError)
			}

			if err != nil && req != nil {
				t.Error("request must be nil in case of error")
			}
		})
	}
}
