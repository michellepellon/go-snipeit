// Package snipeit provides a client for the Snipe-IT Asset Management API.
package snipeit

import (
	"context"
	"math"
	"net/http"
	"sync"
	"time"
)

// RateLimiter defines the interface for rate limiting API requests.
type RateLimiter interface {
	// Wait blocks until a request can be made according to the rate limit.
	Wait(ctx context.Context) error
}

// TokenBucketRateLimiter implements a simple token bucket rate limiter.
type TokenBucketRateLimiter struct {
	tokens         float64
	maxTokens      float64
	tokensPerSec   float64
	lastRefillTime time.Time
	mutex          sync.Mutex
}

// NewTokenBucketRateLimiter creates a new token bucket rate limiter.
//
// requestsPerSecond is the maximum number of requests allowed per second.
// burstSize is the maximum number of requests that can be made in a burst.
func NewTokenBucketRateLimiter(requestsPerSecond float64, burstSize int) *TokenBucketRateLimiter {
	if requestsPerSecond <= 0 {
		requestsPerSecond = float64(defaultMaxRequestsPerSecond)
	}
	if burstSize <= 0 {
		burstSize = defaultBurstSize
	}

	return &TokenBucketRateLimiter{
		tokens:         float64(burstSize),
		maxTokens:      float64(burstSize),
		tokensPerSec:   requestsPerSecond,
		lastRefillTime: time.Now(),
	}
}

// Wait blocks until a token is available or the context is canceled.
func (r *TokenBucketRateLimiter) Wait(ctx context.Context) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(r.lastRefillTime).Seconds()
	r.tokens = math.Min(r.maxTokens, r.tokens+elapsed*r.tokensPerSec)
	r.lastRefillTime = now

	// If we have at least one token, consume it immediately
	if r.tokens >= 1 {
		r.tokens--
		return nil
	}

	// Calculate wait time until next token is available
	waitTime := time.Duration((1.0 - r.tokens) / r.tokensPerSec * float64(time.Second))

	// Create a timer for the wait
	timer := time.NewTimer(waitTime)
	defer timer.Stop()

	// Wait for either the timer to expire or the context to be canceled
	select {
	case <-timer.C:
		// Timer expired, we can make the request
		r.tokens = 0 // Consumed token
		return nil
	case <-ctx.Done():
		// Context was canceled
		return ctx.Err()
	}
}

// RetryPolicy defines how requests should be retried.
type RetryPolicy struct {
	// MaxRetries is the maximum number of times to retry a failed request.
	MaxRetries int

	// RetryableStatusCodes is a map of HTTP status codes that should trigger a retry.
	RetryableStatusCodes map[int]bool

	// InitialBackoff is the initial backoff duration before the first retry.
	InitialBackoff time.Duration

	// MaxBackoff is the maximum backoff duration between retries.
	MaxBackoff time.Duration

	// BackoffMultiplier is the factor by which the backoff increases after each retry.
	BackoffMultiplier float64

	// Jitter is a factor of randomness to add to the backoff to prevent clients
	// from retrying in lockstep. It's a value between 0 and 1, where 0 means no jitter
	// and 1 means the backoff can be anywhere from 0 to the calculated backoff time.
	Jitter float64
}

// DefaultRetryPolicy returns the default retry policy.
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries: defaultMaxRetries,
		RetryableStatusCodes: map[int]bool{
			http.StatusTooManyRequests:     true, // 429
			http.StatusInternalServerError: true, // 500
			http.StatusBadGateway:          true, // 502
			http.StatusServiceUnavailable:  true, // 503
			http.StatusGatewayTimeout:      true, // 504
		},
		InitialBackoff:    defaultInitialBackoff,
		MaxBackoff:        defaultMaxBackoff,
		BackoffMultiplier: defaultBackoffMultiplier,
		Jitter:            defaultJitter,
	}
}

// Default values for rate limiting and retry
const (
	defaultMaxRequestsPerSecond = 10
	defaultBurstSize            = 15
	defaultMaxRetries           = 3
	defaultInitialBackoff       = 1 * time.Second
	defaultMaxBackoff           = 30 * time.Second
	defaultBackoffMultiplier    = 2.0
	defaultJitter               = 0.2
)

// ClientOptions contains options for configuring the Snipe-IT client.
type ClientOptions struct {
	// HTTPClient is the HTTP client to use for making requests.
	// If nil, http.DefaultClient will be used.
	HTTPClient *http.Client

	// RateLimiter controls the rate at which requests are made to the API.
	// If nil, no rate limiting will be applied.
	RateLimiter RateLimiter

	// RetryPolicy defines how failed requests should be retried.
	// If nil, DefaultRetryPolicy will be used.
	RetryPolicy *RetryPolicy

	// DisableRetries, if true, disables automatic retries for failed requests.
	DisableRetries bool
}

// RequestOptions contains options for individual API requests.
type RequestOptions struct {
	// Context is the context for the request.
	// If nil, context.Background() will be used.
	Context context.Context

	// DisableRetries, if true, disables automatic retries for this request,
	// regardless of the client's retry configuration.
	DisableRetries bool
}