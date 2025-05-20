package snipeit

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"
)

func setup() (client *Client, mux *http.ServeMux, serverURL string, teardown func()) {
	// Create a test server
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)

	// Create a client that points to the test server
	client, _ = NewClient(server.URL, "test-token")

	return client, mux, server.URL, server.Close
}

func testMethod(t *testing.T, r *http.Request, expected string) {
	if r.Method != expected {
		t.Errorf("Request method = %v, expected %v", r.Method, expected)
	}
}

func testHeader(t *testing.T, r *http.Request, header string, expected string) {
	if got := r.Header.Get(header); got != expected {
		t.Errorf("Header.Get(%q) = %q, expected %q", header, got, expected)
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		baseURL   string
		token     string
		wantError bool
	}{
		{
			name:      "Valid inputs",
			baseURL:   "https://example.com",
			token:     "valid-token",
			wantError: false,
		},
		{
			name:      "Empty baseURL",
			baseURL:   "",
			token:     "valid-token",
			wantError: true,
		},
		{
			name:      "Empty token",
			baseURL:   "https://example.com",
			token:     "",
			wantError: true,
		},
		{
			name:      "Invalid URL",
			baseURL:   "://invalid-url",
			token:     "valid-token",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(tt.baseURL, tt.token)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("NewClient(%q, %q) expected error, got none", tt.baseURL, tt.token)
				}
				return
			}
			
			if err != nil {
				t.Errorf("NewClient(%q, %q) unexpected error: %v", tt.baseURL, tt.token, err)
				return
			}
			
			if c.client == nil {
				t.Error("NewClient returned nil http.Client")
			}
			
			if c.BaseURL == nil {
				t.Error("NewClient returned nil BaseURL")
			}
			
			if c.token != "Bearer "+tt.token {
				t.Errorf("NewClient token = %q, expected %q", c.token, "Bearer "+tt.token)
			}
			
			// Test trailing slash is added
			if c.BaseURL.Path != "/" && !tt.wantError {
				t.Errorf("NewClient BaseURL.Path = %q, expected to have trailing slash", c.BaseURL.Path)
			}
		})
	}
}

func TestNewClientWithHTTPClient(t *testing.T) {
	customClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	tests := []struct {
		name       string
		httpClient *http.Client
		baseURL    string
		token      string
		wantError  bool
	}{
		{
			name:       "Valid inputs with custom client",
			httpClient: customClient,
			baseURL:    "https://example.com",
			token:      "valid-token",
			wantError:  false,
		},
		{
			name:       "Valid inputs with nil client",
			httpClient: nil,
			baseURL:    "https://example.com",
			token:      "valid-token",
			wantError:  false,
		},
		{
			name:       "Empty baseURL",
			httpClient: customClient,
			baseURL:    "",
			token:      "valid-token",
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClientWithHTTPClient(tt.httpClient, tt.baseURL, tt.token)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("NewClientWithHTTPClient(%v, %q, %q) expected error, got none", 
						tt.httpClient, tt.baseURL, tt.token)
				}
				return
			}
			
			if err != nil {
				t.Errorf("NewClientWithHTTPClient(%v, %q, %q) unexpected error: %v", 
					tt.httpClient, tt.baseURL, tt.token, err)
				return
			}
			
			if c.client == nil {
				t.Error("NewClientWithHTTPClient returned nil http.Client")
			}
			
			// Check if custom client was properly set
			if tt.httpClient != nil && c.client != tt.httpClient {
				t.Errorf("NewClientWithHTTPClient did not set the provided HTTP client")
			}
			
			if c.BaseURL == nil {
				t.Error("NewClientWithHTTPClient returned nil BaseURL")
			}
			
			if c.token != "Bearer "+tt.token {
				t.Errorf("NewClientWithHTTPClient token = %q, expected %q", c.token, "Bearer "+tt.token)
			}
			
			// Test trailing slash is added
			if c.BaseURL.Path != "/" {
				t.Errorf("NewClientWithHTTPClient BaseURL.Path = %q, expected to have trailing slash", c.BaseURL.Path)
			}
		})
	}
}

func TestNewRequest(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	type testBody struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
	}

	tests := []struct {
		name      string
		method    string
		urlStr    string
		body      interface{}
		wantError bool
	}{
		{
			name:      "Valid GET request",
			method:    "GET",
			urlStr:    "api/v1/test",
			body:      nil,
			wantError: false,
		},
		{
			name:      "Valid POST request with body",
			method:    "POST",
			urlStr:    "api/v1/test",
			body:      &testBody{Field1: "test", Field2: 123},
			wantError: false,
		},
		{
			name:      "URL parse error",
			method:    "GET",
			urlStr:    ":%invalid-url",
			body:      nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := client.newRequest(tt.method, tt.urlStr, tt.body)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("newRequest(%q, %q, %v) expected error, got none", tt.method, tt.urlStr, tt.body)
				}
				return
			}
			
			if err != nil {
				t.Errorf("newRequest(%q, %q, %v) unexpected error: %v", tt.method, tt.urlStr, tt.body, err)
				return
			}
			
			// Check method
			if req.Method != tt.method {
				t.Errorf("newRequest() method = %v, expected %v", req.Method, tt.method)
			}
			
			// Check headers
			expectedHeaders := map[string]string{
				"Accept":        "application/json",
				"Content-Type":  "application/json",
				"Authorization": client.token,
			}
			for header, value := range expectedHeaders {
				if got := req.Header.Get(header); got != value {
					t.Errorf("newRequest() header %q = %q, expected %q", header, got, value)
				}
			}
			
			// Check URL
			expectedURL := fmt.Sprintf("%s%s", client.BaseURL, tt.urlStr)
			if req.URL.String() != expectedURL {
				t.Errorf("newRequest() URL = %q, expected %q", req.URL.String(), expectedURL)
			}
		})
	}
}

func TestDo(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	type testResponse struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
	}

	// Test successful request
	mux.HandleFunc("/api/v1/test", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"field1":"test-value","field2":123}`)
	})

	req, _ := client.newRequest("GET", "api/v1/test", nil)
	var response testResponse
	resp, err := client.Do(req, &response)

	if err != nil {
		t.Fatalf("Do() unexpected error: %v", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Do() status = %d, expected %d", resp.StatusCode, http.StatusOK)
	}
	
	expected := testResponse{Field1: "test-value", Field2: 123}
	if !reflect.DeepEqual(response, expected) {
		t.Errorf("Do() response = %+v, expected %+v", response, expected)
	}

	// Test non-2xx response
	mux.HandleFunc("/api/v1/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message":"Bad request error"}`)
	})

	req, _ = client.newRequest("GET", "api/v1/error", nil)
	resp, err = client.Do(req, nil)

	if err == nil {
		t.Fatal("Do() expected error, got none")
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Do() status = %d, expected %d", resp.StatusCode, http.StatusBadRequest)
	}

	errorResponse, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("Do() error type = %T, expected *ErrorResponse", err)
	}

	if errorResponse.Message != "Bad request error" {
		t.Errorf("ErrorResponse.Message = %q, expected %q", errorResponse.Message, "Bad request error")
	}
}

func TestAddOptions(t *testing.T) {
	client, _, _, teardown := setup()
	defer teardown()

	type testOptions struct {
		Field1 string `url:"field1,omitempty"`
		Field2 int    `url:"field2,omitempty"`
		Field3 string `url:"field3"`
		Field4 *int   `url:"field4,omitempty"`
	}

	tests := []struct {
		name           string
		url            string
		options        interface{}
		expectedParams map[string]string
		wantError      bool
	}{
		{
			name:           "All fields with values",
			url:            "api/v1/test",
			options:        &testOptions{Field1: "value1", Field2: 123, Field3: "value3", Field4: new(int)},
			expectedParams: map[string]string{"field1": "value1", "field2": "123", "field3": "value3", "field4": "0"},
			wantError:      false,
		},
		{
			name:           "Omit empty fields",
			url:            "api/v1/test",
			options:        &testOptions{Field3: "value3"},
			expectedParams: map[string]string{"field3": "value3"},
			wantError:      false,
		},
		{
			name:           "Nil options",
			url:            "api/v1/test",
			options:        nil,
			expectedParams: map[string]string{},
			wantError:      false,
		},
		{
			name:           "Invalid URL",
			url:            ":%invalid-url",
			options:        &testOptions{},
			expectedParams: nil,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultURL, err := client.AddOptions(tt.url, tt.options)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("AddOptions(%q, %+v) expected error, got none", tt.url, tt.options)
				}
				return
			}
			
			if err != nil {
				t.Errorf("AddOptions(%q, %+v) unexpected error: %v", tt.url, tt.options, err)
				return
			}
			
			u, _ := url.Parse(resultURL)
			
			params := u.Query()
			if len(params) != len(tt.expectedParams) {
				t.Errorf("AddOptions() params = %v, expected %v", params, tt.expectedParams)
			}
			
			for key, expectedValue := range tt.expectedParams {
				if params.Get(key) != expectedValue {
					t.Errorf("AddOptions() param %q = %q, expected %q", key, params.Get(key), expectedValue)
				}
			}
		})
	}
}