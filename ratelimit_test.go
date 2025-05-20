package snipeit

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTokenBucketRateLimiter(t *testing.T) {
	// Create a rate limiter with 10 requests per second and burst of 5
	limiter := NewTokenBucketRateLimiter(10, 5)

	// Test that we can make 5 requests immediately (due to burst capacity)
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		err := limiter.Wait(ctx)
		if err != nil {
			t.Errorf("Expected no error for request %d, got: %v", i, err)
		}
	}

	// Make one more request which should require waiting
	start := time.Now()
	err := limiter.Wait(ctx)
	elapsed := time.Since(start)
	if err != nil {
		t.Errorf("Expected no error for request after burst, got: %v", err)
	}
	if elapsed < 90*time.Millisecond {
		t.Errorf("Expected to wait at least 100ms, but only waited %v", elapsed)
	}
}

func TestTokenBucketRateLimiterCancellation(t *testing.T) {
	// Create a rate limiter with very slow rate (1 request per 10 seconds)
	limiter := NewTokenBucketRateLimiter(0.1, 1)

	// Use up the burst capacity
	ctx := context.Background()
	err := limiter.Wait(ctx)
	if err != nil {
		t.Errorf("Expected no error for initial request, got: %v", err)
	}

	// Create a cancelable context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// The next request should be cancelled by the context timeout
	err = limiter.Wait(ctx)
	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	} else if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got: %v", err)
	}
}

func TestClientRetry(t *testing.T) {
	// Number of server request attempts
	attempts := 0
	maxAttempts := 3

	// Create a test server that fails for the first 2 attempts
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < maxAttempts {
			// Return a 503 Service Unavailable
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintln(w, `{"status":"error","message":"Service unavailable"}`)
			return
		}
		// Success on the third attempt
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"success","message":"OK"}`)
	}))
	defer server.Close()

	// Create a client with retry enabled and very short backoff for testing
	retryPolicy := DefaultRetryPolicy()
	retryPolicy.InitialBackoff = 10 * time.Millisecond
	retryPolicy.MaxBackoff = 50 * time.Millisecond

	client, err := NewClientWithOptions(server.URL, "test-token", &ClientOptions{
		RetryPolicy: retryPolicy,
	})
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}

	// Make a request that should retry and eventually succeed
	req, err := client.newRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}

	var response Response
	resp, err := client.Do(req, &response)
	if err != nil {
		t.Fatalf("Expected success after retries, got error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	if response.Status != "success" {
		t.Errorf("Expected success status, got %s", response.Status)
	}

	// Verify that the server received all retry attempts
	if attempts != maxAttempts {
		t.Errorf("Expected %d attempts, got %d", maxAttempts, attempts)
	}
}

func TestClientRetryWithDisabled(t *testing.T) {
	// Number of server request attempts
	attempts := 0

	// Create a test server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		// Return a 503 Service Unavailable
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(w, `{"status":"error","message":"Service unavailable"}`)
	}))
	defer server.Close()

	// Create a client with retry disabled
	client, err := NewClientWithOptions(server.URL, "test-token", &ClientOptions{
		DisableRetries: true,
	})
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}

	// Make a request that should not retry
	req, err := client.newRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}

	var response Response
	_, err = client.Do(req, &response)
	if err == nil {
		t.Fatalf("Expected an error but got nil")
	}

	// Verify that the server received exactly one attempt
	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempts)
	}
}

func TestClientRetryAfterHeader(t *testing.T) {
	// Number of server request attempts
	attempts := 0
	maxAttempts := 2

	// Create a test server that returns Retry-After header on first attempt
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < maxAttempts {
			// Return a 429 Too Many Requests with a Retry-After header
			w.Header().Set("Retry-After", "1") // 1 second
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprintln(w, `{"status":"error","message":"Rate limit exceeded"}`)
			return
		}
		// Success on the second attempt
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"success","message":"OK"}`)
	}))
	defer server.Close()

	// Create a client with retry enabled
	client, err := NewClientWithOptions(server.URL, "test-token", &ClientOptions{
		RetryPolicy: &RetryPolicy{
			MaxRetries: 1,
			RetryableStatusCodes: map[int]bool{
				http.StatusTooManyRequests: true,
			},
			InitialBackoff:    500 * time.Millisecond,
			MaxBackoff:        5 * time.Second,
			BackoffMultiplier: 2.0,
			Jitter:            0.2,
		},
	})
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}

	// Use a shorter waiting time for testing purposes
	// In a real app, we'd respect the full Retry-After duration
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Make a request that should retry and eventually succeed
	req, err := client.newRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}
	req = req.WithContext(ctx)

	var response Response
	resp, err := client.Do(req, &response)
	if err != nil {
		t.Fatalf("Expected success after retries, got error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	if response.Status != "success" {
		t.Errorf("Expected success status, got %s", response.Status)
	}

	// Verify that the server received the expected number of attempts
	if attempts != maxAttempts {
		t.Errorf("Expected %d attempts, got %d", maxAttempts, attempts)
	}
}