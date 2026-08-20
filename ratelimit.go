// Package snipeit provides a client for the Snipe-IT Asset Management API.
package snipeit

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"strings"
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

	// RetryMethods limits which HTTP methods may be replayed after a failure
	// the server may already have processed — a 5xx response or a transport
	// error. HTTP 429 is exempt: the request was rejected before processing,
	// so it is retried for every method.
	//
	// If nil, the safe RFC 9110 idempotent methods are used (GET, HEAD,
	// OPTIONS, PUT, DELETE). Add POST or PATCH only if the API calls they
	// carry are idempotent in your usage; replaying them can otherwise create
	// duplicate records.
	RetryMethods map[string]bool
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

// defaultRetryMethods are the methods replayed after a 5xx or transport error
// when RetryPolicy.RetryMethods is nil.
var defaultRetryMethods = map[string]bool{
	http.MethodGet:     true,
	http.MethodHead:    true,
	http.MethodOptions: true,
	http.MethodPut:     true,
	http.MethodDelete:  true,
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

// Logger is an interface for logging HTTP request and response details.
// Implement this interface and set it on ClientOptions to enable debug logging.
type Logger interface {
	// LogRequest is called before each HTTP request is sent.
	// method is the HTTP method, url is the full request URL, and body is the
	// request body (nil for requests with no body).
	LogRequest(method, url string, body []byte)

	// LogResponse is called after each HTTP response is received.
	// method is the HTTP method, url is the full request URL, statusCode is
	// the HTTP status code, and body is the response body.
	LogResponse(method, url string, statusCode int, body []byte)
}

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

	// Logger enables debug logging of HTTP requests and responses.
	// If nil, no logging is performed.
	Logger Logger

	// OnRateLimit, if set, is called with the rate-limit snapshot parsed from
	// every response that carries X-Ratelimit-* headers. Use it to log or
	// export how much of the instance's budget remains. It runs on the
	// request's goroutine, so it must not block.
	OnRateLimit func(RateLimit)
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

// RateLimit is a snapshot of the server's rate-limit state, parsed from the
// response headers Snipe-IT sends on every API request:
//
//	X-Ratelimit-Limit            requests allowed per window
//	X-Ratelimit-Remaining        requests left in the current window
//	X-Ratelimit-Reset            seconds until the window resets
//	X-Ratelimit-Reset-Timestamp  unix timestamp of the reset
//	Retry-After                  seconds to wait (sent with 429 responses)
//
// Snipe-IT Cloud plans differ in their per-minute allowance, so reading these
// headers is the only reliable way for a client to pace itself.
type RateLimit struct {
	// Limit is the number of requests allowed per window (0 if not advertised).
	Limit int

	// Remaining is the number of requests left in the current window.
	Remaining int

	// Reset is the time until the current window resets (0 if not advertised).
	Reset time.Duration

	// ResetAt is the absolute time the window resets, from
	// X-Ratelimit-Reset-Timestamp. Zero if the header is absent or unparseable.
	ResetAt time.Time

	// RetryAfter is the server-requested wait from the Retry-After header.
	// Zero if the header is absent, an HTTP-date, or negative.
	//
	// Snipe-IT sends this header on EVERY response, not only on a 429, where
	// it describes when the current window rolls over rather than asking the
	// client to stop. Only treat it as a demand to wait when StatusCode says
	// the request was actually rejected — see RetryAfterIsBinding.
	RetryAfter time.Duration

	// StatusCode is the status of the response the snapshot came from, or 0
	// when the snapshot was parsed from bare headers.
	StatusCode int

	// Observed is the time the snapshot was taken.
	Observed time.Time

	// Valid reports whether any rate-limit header was present.
	Valid bool
}

// Exhausted reports whether the advertised budget for the current window is
// used up, i.e. further requests will be rejected with HTTP 429 until Reset.
func (r RateLimit) Exhausted() bool { return r.Valid && r.Limit > 0 && r.Remaining <= 0 }

// RetryAfterIsBinding reports whether Retry-After is the server telling the
// client to stop, rather than incidental window information attached to a
// response that succeeded.
func (r RateLimit) RetryAfterIsBinding() bool {
	return r.RetryAfter > 0 && r.StatusCode >= 400
}

// ParseRateLimitResponse extracts the rate-limit state from a response,
// recording its status so callers can tell a throttled response from a
// successful one that merely carries window information.
func ParseRateLimitResponse(resp *http.Response) (RateLimit, bool) {
	if resp == nil {
		return RateLimit{}, false
	}
	rl, ok := ParseRateLimit(resp.Header)
	rl.StatusCode = resp.StatusCode
	return rl, ok
}

// ParseRateLimit extracts the rate-limit state from a response's headers. The
// returned bool (mirrored by RateLimit.Valid) is false when the response
// carried none of the headers, e.g. a Snipe-IT instance with rate limiting
// disabled or a non-Snipe intermediary's error page.
//
// Prefer ParseRateLimitResponse when the response is at hand: a Retry-After
// value cannot be interpreted without knowing the status it arrived with.
func ParseRateLimit(h http.Header) (RateLimit, bool) {
	rl := RateLimit{Observed: time.Now()}
	if v, ok := headerInt(h, "X-Ratelimit-Limit"); ok {
		rl.Limit, rl.Valid = v, true
	}
	if v, ok := headerInt(h, "X-Ratelimit-Remaining"); ok {
		rl.Remaining, rl.Valid = v, true
	}
	if v, ok := headerInt(h, "X-Ratelimit-Reset"); ok && v >= 0 {
		rl.Reset, rl.Valid = time.Duration(v)*time.Second, true
	}
	if v, ok := headerInt(h, "X-Ratelimit-Reset-Timestamp"); ok && v > 0 {
		rl.ResetAt, rl.Valid = time.Unix(int64(v), 0), true
	}
	if d, ok := retryAfter(h); ok {
		rl.RetryAfter, rl.Valid = d, true
	}
	return rl, rl.Valid
}

// headerInt reads a header as a base-10 integer.
func headerInt(h http.Header, name string) (int, bool) {
	raw := strings.TrimSpace(h.Get(name))
	if raw == "" {
		return 0, false
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	return v, true
}

// retryAfter parses a Retry-After header in either of its RFC 9110 forms:
// delta-seconds (Snipe-IT sends this) or an HTTP-date. A negative delta or a
// date already in the past yields zero. The bool reports whether a usable
// value was present, so callers can distinguish "wait 0s" from "no header".
func retryAfter(h http.Header) (time.Duration, bool) {
	raw := strings.TrimSpace(h.Get("Retry-After"))
	if raw == "" {
		return 0, false
	}
	if secs, err := strconv.Atoi(raw); err == nil {
		if secs < 0 {
			return 0, true
		}
		return time.Duration(secs) * time.Second, true
	}
	if t, err := http.ParseTime(raw); err == nil {
		if d := time.Until(t); d > 0 {
			return d, true
		}
		return 0, true
	}
	return 0, false
}

// RateLimitObserver is implemented by rate limiters that adapt to the limits
// the server advertises. A client feeds every response's rate-limit snapshot
// to its limiter when the limiter implements this interface.
type RateLimitObserver interface {
	Observe(RateLimit)
}

// AdaptiveRateLimiter paces requests at a configured rate and then tightens
// that rate using the server's own X-Ratelimit-* headers, so a client stays
// under the instance's budget instead of discovering it through 429s.
//
// It behaves as a token bucket at the configured rate until the first response
// is observed. From then on the pace is the lower of the configured rate and
// the rate that spreads the remaining requests evenly over the rest of the
// window. When the window's budget is exhausted (or the server sends a
// Retry-After), Wait blocks until the window resets.
type AdaptiveRateLimiter struct {
	mu sync.Mutex

	baseRate  float64 // configured requests per second
	rate      float64 // current effective requests per second
	tokens    float64
	maxTokens float64
	lastFill  time.Time

	// notBefore holds requests until the current window resets, set when the
	// budget is exhausted or the server asks for a Retry-After wait.
	notBefore time.Time

	// last is the most recent observation, exposed via Snapshot for callers
	// that want to log or export the server's view.
	last RateLimit

	// now is time.Now, overridable in tests.
	now func() time.Time
}

// NewAdaptiveRateLimiter creates a limiter that starts at requestsPerSecond
// with the given burst size and adapts to the server's advertised limits.
//
// Pick requestsPerSecond from the instance's plan (Snipe-IT Cloud publishes a
// per-minute allowance); the limiter only ever paces slower than this value,
// never faster, so a conservative starting point is safe.
func NewAdaptiveRateLimiter(requestsPerSecond float64, burstSize int) *AdaptiveRateLimiter {
	if requestsPerSecond <= 0 {
		requestsPerSecond = float64(defaultMaxRequestsPerSecond)
	}
	if burstSize <= 0 {
		burstSize = defaultBurstSize
	}
	return &AdaptiveRateLimiter{
		baseRate:  requestsPerSecond,
		rate:      requestsPerSecond,
		tokens:    float64(burstSize),
		maxTokens: float64(burstSize),
		lastFill:  time.Now(),
		now:       time.Now,
	}
}

// Wait blocks until the next request may be made or ctx is canceled.
func (a *AdaptiveRateLimiter) Wait(ctx context.Context) error {
	for {
		wait := a.reserve()
		if wait <= 0 {
			return nil
		}
		timer := time.NewTimer(wait)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}
	}
}

// reserve consumes a token if one is available and reports how long the caller
// must wait before trying again.
func (a *AdaptiveRateLimiter) reserve() time.Duration {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := a.now()
	if now.Before(a.notBefore) {
		return a.notBefore.Sub(now)
	}

	elapsed := now.Sub(a.lastFill).Seconds()
	if elapsed > 0 {
		a.tokens = math.Min(a.maxTokens, a.tokens+elapsed*a.rate)
		a.lastFill = now
	}
	if a.tokens >= 1 {
		a.tokens--
		return 0
	}
	return time.Duration((1 - a.tokens) / a.rate * float64(time.Second))
}

// Observe folds a response's rate-limit snapshot into the pace. It is safe to
// call concurrently and ignores snapshots with no usable headers.
func (a *AdaptiveRateLimiter) Observe(rl RateLimit) {
	if !rl.Valid {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.last = rl
	now := a.now()

	// A server-requested wait wins when the server actually rejected the
	// request: it is the only signal reflecting state this client cannot see,
	// such as another token spending the same budget. On a successful response
	// the same header just describes the window, so holding for it would idle
	// the client for a full window after every request.
	if rl.RetryAfterIsBinding() {
		a.holdUntil(now.Add(rl.RetryAfter))
	}

	reset := rl.Reset
	if !rl.ResetAt.IsZero() {
		// Prefer the absolute timestamp; it stays correct while a request is
		// in flight. Ignore it when it has already passed.
		if d := rl.ResetAt.Sub(now); d > 0 {
			reset = d
		}
	}

	if rl.Exhausted() && reset > 0 {
		a.holdUntil(now.Add(reset))
		return
	}
	if rl.Limit <= 0 || reset <= 0 || rl.Remaining <= 0 {
		return
	}

	// Spread what is left over the rest of the window, but never speed up past
	// the configured rate.
	a.rate = math.Min(a.baseRate, float64(rl.Remaining)/reset.Seconds())
}

// holdUntil blocks new requests until t and drains the burst so the window
// does not reopen with a full bucket.
func (a *AdaptiveRateLimiter) holdUntil(t time.Time) {
	if t.After(a.notBefore) {
		a.notBefore = t
	}
	a.tokens = 0
	a.lastFill = t
	a.rate = a.baseRate
}

// Snapshot returns the most recent observation and the pace the limiter is
// currently enforcing, for logging and metrics.
func (a *AdaptiveRateLimiter) Snapshot() (RateLimit, float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.last, a.rate
}

// RateLimitPreset is a Snipe-IT Cloud plan's published request allowance,
// so callers can pick a plan by name instead of hand-computing a rate.
type RateLimitPreset struct {
	// Name is the plan's config-file identifier, e.g. "small_business".
	Name string

	// RequestsPerMinute is the plan's allowance, or 0 when the plan is
	// unmetered.
	RequestsPerMinute int

	// Burst is how many requests may be made back to back before pacing
	// kicks in.
	Burst int
}

// Snipe-IT Cloud plan allowances. A preset only sets the ceiling: the limiter
// it builds still tightens its pace from the server's X-Ratelimit-* headers,
// so a plan change or a shared token cannot push the client into 429s.
var (
	// PresetBasic is the Basic plan: 120 requests per minute.
	PresetBasic = RateLimitPreset{Name: "basic", RequestsPerMinute: 120, Burst: 5}

	// PresetSmallBusiness is the Small Business plan: 240 requests per minute.
	PresetSmallBusiness = RateLimitPreset{Name: "small_business", RequestsPerMinute: 240, Burst: 10}

	// PresetDedicated is a dedicated instance, which enforces no limit.
	PresetDedicated = RateLimitPreset{Name: "dedicated"}
)

// Limiter builds the RateLimiter for the preset, ready to hand to
// ClientOptions.RateLimiter. It returns nil for an unmetered preset, which the
// client treats as "no rate limiting".
func (p RateLimitPreset) Limiter() RateLimiter {
	if p.RequestsPerMinute <= 0 {
		return nil
	}
	return NewAdaptiveRateLimiter(float64(p.RequestsPerMinute)/60, p.Burst)
}

// PresetByName looks up a plan preset by its config-file name, matching
// case-insensitively and accepting "-" or nothing in place of the "_".
func PresetByName(name string) (RateLimitPreset, bool) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "basic":
		return PresetBasic, true
	case "small_business", "small-business", "smallbusiness":
		return PresetSmallBusiness, true
	case "dedicated":
		return PresetDedicated, true
	}
	return RateLimitPreset{}, false
}
