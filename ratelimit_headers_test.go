package snipeit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestParseRateLimit(t *testing.T) {
	h := http.Header{}
	h.Set("X-Ratelimit-Limit", "300")
	h.Set("X-Ratelimit-Remaining", "299")
	h.Set("Retry-After", "58")
	h.Set("X-Ratelimit-Reset", "58")
	h.Set("X-Ratelimit-Reset-Timestamp", "1787241815")

	rl, ok := ParseRateLimit(h)
	if !ok || !rl.Valid {
		t.Fatal("expected the headers to parse")
	}
	if rl.Limit != 300 || rl.Remaining != 299 {
		t.Errorf("limit/remaining = %d/%d, want 300/299", rl.Limit, rl.Remaining)
	}
	if rl.Reset != 58*time.Second || rl.RetryAfter != 58*time.Second {
		t.Errorf("reset/retry-after = %s/%s, want 58s/58s", rl.Reset, rl.RetryAfter)
	}
	if rl.ResetAt.Unix() != 1787241815 {
		t.Errorf("reset-at = %s", rl.ResetAt)
	}
	if rl.Exhausted() {
		t.Error("299 of 300 remaining is not exhausted")
	}

	if _, ok := ParseRateLimit(http.Header{}); ok {
		t.Error("an empty header set must not report a rate limit")
	}
}

func TestParseRateLimitExhausted(t *testing.T) {
	h := http.Header{}
	h.Set("X-Ratelimit-Limit", "240")
	h.Set("X-Ratelimit-Remaining", "0")
	rl, _ := ParseRateLimit(h)
	if !rl.Exhausted() {
		t.Error("0 remaining must report exhausted")
	}
}

func TestRetryAfterHTTPDate(t *testing.T) {
	h := http.Header{}
	h.Set("Retry-After", time.Now().Add(30*time.Second).UTC().Format(http.TimeFormat))
	d, ok := retryAfter(h)
	if !ok || d <= 0 || d > 31*time.Second {
		t.Errorf("http-date Retry-After = %s, ok=%v", d, ok)
	}

	h.Set("Retry-After", "-5")
	if d, ok := retryAfter(h); !ok || d != 0 {
		t.Errorf("negative Retry-After = %s, ok=%v; want 0, true", d, ok)
	}

	h.Set("Retry-After", "soon")
	if _, ok := retryAfter(h); ok {
		t.Error("an unparseable Retry-After must report absent")
	}
}

// The client must record every response's budget, including a 429's, and hand
// it to both the callback and an adaptive limiter.
func TestClientObservesRateLimitHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Ratelimit-Limit", "240")
		w.Header().Set("X-Ratelimit-Remaining", "12")
		w.Header().Set("X-Ratelimit-Reset", "30")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"total":0,"rows":[]}`))
	}))
	defer srv.Close()

	var seen int32
	limiter := NewAdaptiveRateLimiter(100, 5)
	c, err := NewClientWithOptions(srv.URL, "tok", &ClientOptions{
		RateLimiter: limiter,
		OnRateLimit: func(RateLimit) { atomic.AddInt32(&seen, 1) },
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := c.Assets.List(&ListOptions{Limit: 1}); err != nil {
		t.Fatal(err)
	}
	if atomic.LoadInt32(&seen) != 1 {
		t.Errorf("OnRateLimit called %d times, want 1", seen)
	}
	if rl := c.RateLimit(); rl.Limit != 240 || rl.Remaining != 12 {
		t.Errorf("client rate limit = %+v", rl)
	}
	// 12 requests left over 30s is 0.4/s, well under the configured 100/s.
	if _, rate := limiter.Snapshot(); rate > 0.5 {
		t.Errorf("limiter rate = %.2f/s, want it tightened to ~0.4/s", rate)
	}
}

func TestAdaptiveRateLimiterHoldsWhenExhausted(t *testing.T) {
	now := time.Now()
	l := NewAdaptiveRateLimiter(10, 5)
	l.now = func() time.Time { return now }

	l.Observe(RateLimit{Valid: true, Limit: 240, Remaining: 0, Reset: 20 * time.Second})
	if wait := l.reserve(); wait < 19*time.Second {
		t.Fatalf("exhausted budget must hold until reset, waited %s", wait)
	}

	// After the window resets the limiter runs at its configured pace again.
	now = now.Add(21 * time.Second)
	if wait := l.reserve(); wait != 0 {
		t.Fatalf("wait after reset = %s, want 0", wait)
	}
}

func TestAdaptiveRateLimiterNeverExceedsBaseRate(t *testing.T) {
	l := NewAdaptiveRateLimiter(4, 1)
	l.Observe(RateLimit{Valid: true, Limit: 10000, Remaining: 9999, Reset: time.Second})
	if _, rate := l.Snapshot(); rate != 4 {
		t.Errorf("rate = %.2f, want it capped at the configured 4/s", rate)
	}
}

func TestAdaptiveRateLimiterWaitRespectsContext(t *testing.T) {
	l := NewAdaptiveRateLimiter(10, 1)
	l.Observe(RateLimit{Valid: true, Limit: 240, Remaining: 0, Reset: time.Hour})
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := l.Wait(ctx); err == nil {
		t.Fatal("Wait must return the context error instead of holding for the window")
	}
}

// A 5xx or transport failure must not replay a POST, which may already have
// created a record; a 429 is safe to repeat for any method.
func TestRetryMethodIdempotency(t *testing.T) {
	policy := DefaultRetryPolicy()
	c := &Client{}

	post := httptest.NewRequest(http.MethodPost, "/api/v1/hardware", nil)
	get := httptest.NewRequest(http.MethodGet, "/api/v1/hardware", nil)
	fail := &http.Response{StatusCode: http.StatusInternalServerError, Header: http.Header{}}

	if retry, _, _ := c.shouldRetry(post, fail, nil, policy); retry {
		t.Error("POST must not be replayed after a 500")
	}
	if retry, _, _ := c.shouldRetry(get, fail, nil, policy); !retry {
		t.Error("GET must be retried after a 500")
	}

	limited := &http.Response{StatusCode: http.StatusTooManyRequests, Header: http.Header{}}
	limited.Header.Set("Retry-After", "3")
	retry, wait, _ := c.shouldRetry(post, limited, nil, policy)
	if !retry || wait != 3*time.Second {
		t.Errorf("429 POST: retry=%v wait=%s, want true/3s", retry, wait)
	}

	// Retry-After is clamped to MaxBackoff.
	limited.Header.Set("Retry-After", "86400")
	if _, wait, _ := c.shouldRetry(get, limited, nil, policy); wait != policy.MaxBackoff {
		t.Errorf("clamped wait = %s, want %s", wait, policy.MaxBackoff)
	}

	// An explicit "Retry-After: 0" means retry now, not "fall back to backoff".
	zero := &http.Response{StatusCode: http.StatusTooManyRequests, Header: http.Header{}}
	zero.Header.Set("Retry-After", "0")
	retry, wait, directed := c.shouldRetry(get, zero, nil, policy)
	if !retry || wait != 0 || !directed {
		t.Errorf("Retry-After 0: retry=%v wait=%s serverDirected=%v, want true/0s/true", retry, wait, directed)
	}

	// With no Retry-After, a 429 falls back to the reset window.
	reset := &http.Response{StatusCode: http.StatusTooManyRequests, Header: http.Header{}}
	reset.Header.Set("X-Ratelimit-Reset", "7")
	if _, wait, _ := c.shouldRetry(get, reset, nil, policy); wait != 7*time.Second {
		t.Errorf("reset-based wait = %s, want 7s", wait)
	}
}

// Snipe-IT sends Retry-After on EVERY response, including 200s, where it means
// "when the current window rolls over" — not "stop until then". Honoring it
// unconditionally parks the limiter for a full window after each successful
// request, turning a paged read into one request per minute.
func TestAdaptiveRateLimiterIgnoresRetryAfterOnSuccess(t *testing.T) {
	now := time.Now()
	l := NewAdaptiveRateLimiter(4, 5)
	l.now = func() time.Time { return now }

	l.Observe(RateLimit{
		Valid: true, StatusCode: http.StatusOK,
		Limit: 240, Remaining: 239, Reset: 59 * time.Second, RetryAfter: 58 * time.Second,
	})
	if wait := l.reserve(); wait != 0 {
		t.Fatalf("a 200 response with Retry-After parked the limiter for %s", wait)
	}

	// A 429's Retry-After still holds: that one really is "do not send yet".
	l.Observe(RateLimit{
		Valid: true, StatusCode: http.StatusTooManyRequests,
		Limit: 240, Remaining: 0, Reset: 30 * time.Second, RetryAfter: 30 * time.Second,
	})
	if wait := l.reserve(); wait < 29*time.Second {
		t.Fatalf("429 Retry-After must hold, waited %s", wait)
	}
}

// The client records the status alongside the headers, so the limiter can tell
// a throttled response from a successful one.
func TestObservedRateLimitCarriesStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Ratelimit-Limit", "240")
		w.Header().Set("X-Ratelimit-Remaining", "239")
		w.Header().Set("Retry-After", "58")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"total":0,"rows":[]}`))
	}))
	defer srv.Close()

	limiter := NewAdaptiveRateLimiter(4, 5)
	c, err := NewClientWithOptions(srv.URL, "tok", &ClientOptions{RateLimiter: limiter})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := c.Assets.List(&ListOptions{Limit: 1}); err != nil {
		t.Fatal(err)
	}
	if got := c.RateLimit().StatusCode; got != http.StatusOK {
		t.Errorf("observed status = %d, want 200", got)
	}
	if wait := limiter.reserve(); wait != 0 {
		t.Errorf("a successful response held the limiter for %s", wait)
	}
}
