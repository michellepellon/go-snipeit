// Package snipeit provides a Go client for interacting with the Snipe-IT Asset Management API.
//
// Snipe-IT is an open-source IT asset management system that allows 
// for tracking hardware, software, and accessories. This client library
// provides a simple interface to interact with the Snipe-IT REST API.
//
// Usage:
//
//	client, err := snipeit.NewClient("https://your-snipeit-instance.com", "your-api-token")
//	if err != nil {
//	    log.Fatalf("Error creating client: %v", err)
//	}
//
//	// List assets
//	assets, _, err := client.Assets.List(nil)
//	if err != nil {
//	    log.Fatalf("Error listing assets: %v", err)
//	}
//
//	// Get a specific asset
//	asset, _, err := client.Assets.Get(1)
//	if err != nil {
//	    log.Fatalf("Error getting asset: %v", err)
//	}
//
// API documentation: https://snipe-it.readme.io/reference/api-overview
package snipeit

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "math/rand"
    "net/http"
    "net/url"
    "reflect"
    "strings"
    "sync"
    "time"

    "github.com/google/go-querystring/query"
)

// Client manages communication with the Snipe-IT API.
//
// Each service of the Snipe-IT API is exposed as a field on the Client struct.
// For example, to access the assets endpoint, use client.Assets.
type Client struct {
    // HTTP client used to communicate with the API
    client  *http.Client
    
    // Snipe-IT API personal token with "Bearer " prefix
    token   string          

    // Base URL for API requests
    BaseURL *url.URL

    // Services for different parts of the Snipe-IT API
    // Assets is the service for interacting with the assets endpoint
    Assets *AssetsService

    // Locations is the service for interacting with the locations endpoint
    Locations *LocationsService

    // StatusLabels is the service for interacting with the status labels endpoint
    StatusLabels *StatusLabelsService

    // Fields is the service for interacting with the custom fields endpoint
    Fields *FieldsService

    // Fieldsets is the service for interacting with the custom fieldsets endpoint
    Fieldsets *FieldsetsService

    // Models is the service for interacting with the models endpoint
    Models *ModelsService

    // Users is the service for interacting with the users endpoint
    Users *UsersService

    // Suppliers is the service for interacting with the suppliers endpoint
    Suppliers *SuppliersService

    // Categories is the service for interacting with the categories endpoint
    Categories *CategoriesService

    // Manufacturers is the service for interacting with the manufacturers endpoint
    Manufacturers *ManufacturersService

    // Rate limiter for controlling request frequency
    rateLimiter RateLimiter
    
    // Retry policy for handling failed requests
    retryPolicy *RetryPolicy
    
    // DisableRetries, if true, disables automatic retries for failed requests
    disableRetries bool

    // Logger for debug logging of requests and responses
    logger Logger

    // onRateLimit is called with every response's rate-limit snapshot
    onRateLimit func(RateLimit)

    // rateLimit holds the most recent rate-limit snapshot, guarded by rateLimitMu
    rateLimitMu sync.RWMutex
    rateLimit   RateLimit
}

// RateLimit returns the rate-limit state advertised by the most recent
// response, so callers can log how much of the instance's budget is left or
// pause their own work before the client has to. RateLimit.Valid is false
// until a response carrying X-Ratelimit-* headers has been received.
func (c *Client) RateLimit() RateLimit {
    c.rateLimitMu.RLock()
    defer c.rateLimitMu.RUnlock()
    return c.rateLimit
}

// observeRateLimit records a response's rate-limit headers, feeds them to an
// adaptive rate limiter, and invokes the OnRateLimit callback.
func (c *Client) observeRateLimit(resp *http.Response) RateLimit {
    if resp == nil {
        return RateLimit{}
    }
    rl, ok := ParseRateLimitResponse(resp)
    if !ok {
        return rl
    }
    c.rateLimitMu.Lock()
    c.rateLimit = rl
    c.rateLimitMu.Unlock()

    if observer, isObserver := c.rateLimiter.(RateLimitObserver); isObserver {
        observer.Observe(rl)
    }
    if c.onRateLimit != nil {
        c.onRateLimit(rl)
    }
    return rl
}

// NewClient returns a new Snipe-IT API client.
//
// baseURL is the base URL of your Snipe-IT instance (e.g., "https://assets.example.com").
// token is your Snipe-IT API token, which can be generated in the Snipe-IT web interface
// under Admin > API Keys.
//
// This function uses the default http.Client. If you need to customize the HTTP client,
// use NewClientWithHTTPClient instead.
//
// If baseURL does not have a trailing slash, one is added automatically.
//
// Returns an error if baseURL is invalid or if either baseURL or token is empty.
func NewClient(baseURL, token string) (*Client, error) {
    return NewClientWithOptions(baseURL, token, nil)
}

// NewClientWithHTTPClient returns a new Snipe-IT API client using the provided HTTP client.
//
// httpClient is the HTTP client to use for making API requests.
// baseURL is the base URL of your Snipe-IT instance.
// token is your Snipe-IT API token.
//
// This function allows you to customize the HTTP client used by the Snipe-IT client,
// which is useful for setting custom timeouts, transport options, or proxies.
//
// If baseURL does not have a trailing slash, one is added automatically.
//
// Returns an error if baseURL is invalid or if either baseURL or token is empty.
func NewClientWithHTTPClient(httpClient *http.Client, baseURL, token string) (*Client, error) {
    options := &ClientOptions{
        HTTPClient: httpClient,
    }
    return NewClientWithOptions(baseURL, token, options)
}

// NewClientWithOptions returns a new Snipe-IT API client with advanced configuration options.
//
// baseURL is the base URL of your Snipe-IT instance (e.g., "https://assets.example.com").
// token is your Snipe-IT API token, which can be generated in the Snipe-IT web interface.
// options allows for configuring rate limiting, retries, and HTTP client settings.
//
// If options is nil, default settings will be used.
// If options.HTTPClient is nil, http.DefaultClient will be used.
// If options.RateLimiter is nil, no rate limiting will be applied.
// If options.RetryPolicy is nil but options.DisableRetries is false, DefaultRetryPolicy will be used.
//
// If baseURL does not have a trailing slash, one is added automatically.
//
// Returns an error if baseURL is invalid or if either baseURL or token is empty.
func NewClientWithOptions(baseURL, token string, options *ClientOptions) (*Client, error) {
    if baseURL == "" {
        return nil, errors.New("a baseURL must be provided")
    }

    if token == "" {
        return nil, errors.New("a token must be provided")
    }

    baseEndpoint, err := url.Parse(baseURL)
    if err != nil {
        return nil, err
    }
    if !strings.HasSuffix(baseEndpoint.Path, "/") {
        baseEndpoint.Path += "/"
    }

    c := new(Client)
    
    if options == nil {
        options = &ClientOptions{}
    }
    
    c.client = options.HTTPClient
    if c.client == nil {
        c.client = &http.Client{}
    }
    
    c.token = "Bearer " + token
    c.BaseURL = baseEndpoint
    
    // Configure rate limiting
    c.rateLimiter = options.RateLimiter
    
    // Configure retry policy
    c.disableRetries = options.DisableRetries
    if !c.disableRetries && options.RetryPolicy == nil {
        c.retryPolicy = DefaultRetryPolicy()
    } else {
        c.retryPolicy = options.RetryPolicy
    }
    
    // Configure logger
    c.logger = options.Logger

    // Configure rate-limit observation
    c.onRateLimit = options.OnRateLimit

    // Initialize services
    c.Assets = &AssetsService{client: c}
    c.Locations = &LocationsService{client: c}
    c.StatusLabels = &StatusLabelsService{client: c}
    c.Fields = &FieldsService{client: c}
    c.Fieldsets = &FieldsetsService{client: c}
    c.Models = &ModelsService{client: c}
    c.Users = &UsersService{client: c}
    c.Suppliers = &SuppliersService{client: c}
    c.Categories = &CategoriesService{client: c}
    c.Manufacturers = &ManufacturersService{client: c}

    return c, nil
}

// DoWithOptions sends an API request with the provided request options and returns the API response.
//
// req is the HTTP request to send.
// v is the destination into which the response JSON will be unmarshaled.
// opts are optional settings for this specific request.
//
// If opts is nil, the client's default options will be used.
// If opts.Context is nil, the request's context will be used.
//
// If the response status code is not in the 2xx range, an ErrorResponse is returned.
// Otherwise, if v is not nil, the response body is JSON decoded into v.
//
// The provided request and returned response are for debugging purposes only and
// should not be directly modified.
func (c *Client) DoWithOptions(req *http.Request, v interface{}, opts *RequestOptions) (*http.Response, error) {
    ctx := req.Context()
    if opts != nil && opts.Context != nil {
        ctx = opts.Context
    }
    
    req = req.WithContext(ctx)
    
    // Rate limiting is applied per attempt inside doOnce, so retries are paced
    // too instead of bursting past the limiter.

    // Determine if retries are enabled for this request
    disableRetries := c.disableRetries
    if opts != nil && opts.DisableRetries {
        disableRetries = true
    }
    
    // If retries are disabled or no retry policy is set, just make a single request
    if disableRetries || c.retryPolicy == nil {
        return c.doOnce(ctx, req, v)
    }
    
    // Initialize retry variables
    var resp *http.Response
    var err error
    var shouldRetry bool
    var retryAfter time.Duration
    var serverDirected bool
    
    retryPolicy := c.retryPolicy
    backoff := retryPolicy.InitialBackoff
    
    // Make the initial request
    resp, err = c.doOnce(ctx, req, v)
    
    // Retry loop
    for retries := 0; retries < retryPolicy.MaxRetries; retries++ {
        // Check if we should retry
        shouldRetry, retryAfter, serverDirected = c.shouldRetry(req, resp, err, retryPolicy)
        if !shouldRetry {
            break
        }
        
        // Log retry attempt (could use a logger here instead of fmt)
        //fmt.Printf("Retrying request to %s after error: %v (attempt %d/%d)\n", 
        //    req.URL.String(), err, retries+1, retryPolicy.MaxRetries)
        
        // Wait before retrying
        if serverDirected {
            // Use the Retry-After header value
            waitTime := retryAfter
            select {
            case <-ctx.Done():
                return resp, ctx.Err()
            case <-time.After(waitTime):
                // Continue with retry
            }
        } else {
            // Calculate backoff with jitter
            jitterRange := backoff.Seconds() * retryPolicy.Jitter
            jitter := time.Duration(rand.Float64() * jitterRange * float64(time.Second))
            waitTime := backoff - jitter
            
            select {
            case <-ctx.Done():
                return resp, ctx.Err()
            case <-time.After(waitTime):
                // Continue with retry
            }
            
            // Increase backoff for next time
            backoff = time.Duration(float64(backoff) * retryPolicy.BackoffMultiplier)
            if backoff > retryPolicy.MaxBackoff {
                backoff = retryPolicy.MaxBackoff
            }
        }
        
        // Create a new request for each retry to ensure a fresh request
        retryReq := req.Clone(ctx)
        
        // If the body was consumed in the previous attempt, we need to recreate it
        if req.GetBody != nil {
            body, err := req.GetBody()
            if err != nil {
                return resp, err
            }
            retryReq.Body = body
        }
        
        // Make the retry request
        resp, err = c.doOnce(ctx, retryReq, v)
    }
    
    return resp, err
}

// doOnce performs a single API request without any retry logic.
func (c *Client) doOnce(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
    // Log request if logger is configured
    if c.logger != nil {
        var reqBody []byte
        if req.Body != nil {
            reqBody, _ = io.ReadAll(req.Body)
            req.Body = io.NopCloser(bytes.NewReader(reqBody))
        }
        c.logger.LogRequest(req.Method, req.URL.String(), reqBody)
    }

    // Apply rate limiting if configured. This sits in doOnce so every attempt
    // — including retries — is paced.
    if c.rateLimiter != nil {
        if err := c.rateLimiter.Wait(ctx); err != nil {
            return nil, err
        }
    }

    resp, err := c.client.Do(req)
    if resp != nil {
        // Record the server's advertised budget before anything can return,
        // so a 429 still updates the limiter and the OnRateLimit callback.
        c.observeRateLimit(resp)
    }
    if err != nil {
        // If the error is due to context cancellation or deadline exceeded,
        // return that specific error
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
        }
        return nil, err
    }
    defer resp.Body.Close()

    // Read the full response body so we can log it and still decode
    respBody, readErr := io.ReadAll(resp.Body)
    if readErr != nil {
        return resp, readErr
    }

    // Log response if logger is configured
    if c.logger != nil {
        c.logger.LogResponse(req.Method, req.URL.String(), resp.StatusCode, respBody)
    }

    // If StatusCode is not in the 200 range, something went wrong
    if sc := resp.StatusCode; 200 > sc || sc > 299 {
        errorResponse := &ErrorResponse{Response: resp}
        if len(respBody) > 0 {
            json.Unmarshal(respBody, errorResponse)
        }
        return resp, errorResponse
    }

    if v != nil {
        if w, ok := v.(io.Writer); ok {
            _, err = w.Write(respBody)
        } else {
            decErr := json.Unmarshal(respBody, v)
            if decErr != nil && len(respBody) > 0 {
                err = decErr
            }
        }
    }

    return resp, err
}

// shouldRetry determines if a request should be retried based on the response, error, and retry policy.
// shouldRetry reports whether to retry, how long the server asked the client to
// wait, and whether that wait came from the server at all — a server-directed
// wait of zero means "retry now" and must not be mistaken for "no guidance".
func (c *Client) shouldRetry(req *http.Request, resp *http.Response, err error, policy *RetryPolicy) (bool, time.Duration, bool) {
    // Don't retry if there's no error and the response is in the 2xx range
    if err == nil && resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
        return false, 0, false
    }

    if resp != nil && policy.RetryableStatusCodes[resp.StatusCode] {
        // A 429 was rejected before the request was processed, so it is always
        // safe to repeat regardless of method.
        if resp.StatusCode != http.StatusTooManyRequests && !policy.retryableMethod(req) {
            return false, 0, false
        }
        wait, ok := retryWait(resp)
        return true, policy.capWait(wait), ok
    }

    // Retry network errors, except for context cancellation. A transport error
    // gives no proof the server did not process the request, so non-idempotent
    // methods are not replayed.
    if err != nil {
        if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
            return false, 0, false
        }
        return policy.retryableMethod(req), 0, false
    }

    return false, 0, false
}

// retryWait reports how long the server asked the client to wait, preferring
// Retry-After and falling back to the time until the rate-limit window resets.
// The bool reports whether the server gave any guidance at all, so an explicit
// "Retry-After: 0" retries immediately instead of falling back to the backoff.
func retryWait(resp *http.Response) (time.Duration, bool) {
    rl, ok := ParseRateLimitResponse(resp)
    if !ok {
        return 0, false
    }
    if resp.Header.Get("Retry-After") != "" {
        return rl.RetryAfter, true
    }
    if resp.StatusCode == http.StatusTooManyRequests && rl.Valid {
        if !rl.ResetAt.IsZero() {
            if d := time.Until(rl.ResetAt); d > 0 {
                return d, true
            }
        }
        if rl.Reset > 0 {
            return rl.Reset, true
        }
    }
    return 0, false
}

// retryableMethod reports whether a request may be replayed after a failure
// that might have been processed by the server.
func (p *RetryPolicy) retryableMethod(req *http.Request) bool {
    if req == nil {
        return false
    }
    methods := p.RetryMethods
    if methods == nil {
        methods = defaultRetryMethods
    }
    return methods[strings.ToUpper(req.Method)]
}

// capWait clamps a server-provided wait to MaxBackoff so one large (or
// malformed-but-numeric) Retry-After cannot stall a client for hours.
func (p *RetryPolicy) capWait(d time.Duration) time.Duration {
    if p.MaxBackoff > 0 && d > p.MaxBackoff {
        return p.MaxBackoff
    }
    return d
}

// newRequest creates an API request.
//
// method is the HTTP method (GET, POST, PUT, DELETE, etc.).
// urlStr is the URL path relative to the BaseURL (e.g., "api/v1/hardware").
// body is the request body to be encoded to JSON (nil for GET requests).
//
// The URL is resolved relative to the BaseURL of the Client.
// If the provided urlStr has a leading slash, it will be trimmed.
// The resulting request will include the proper authentication headers.
func (c *Client) newRequest(method, urlStr string, body interface{}) (*http.Request, error) {
    return c.newRequestWithContext(context.Background(), method, urlStr, body)
}

// NewRequest creates an API request. This is an exported wrapper around newRequest
// for use by callers that need to make direct API calls beyond the built-in services.
func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
    return c.newRequest(method, urlStr, body)
}

// newRequestWithContext creates an API request with the provided context.
//
// ctx is the context for the request. A context with a deadline or timeout
// is especially useful for operations that may take longer than expected.
// method is the HTTP method (GET, POST, PUT, DELETE, etc.).
// urlStr is the URL path relative to the BaseURL (e.g., "api/v1/hardware").
// body is the request body to be encoded to JSON (nil for GET requests).
//
// The URL is resolved relative to the BaseURL of the Client.
// If the provided urlStr has a leading slash, it will be trimmed.
// The resulting request will include the proper authentication headers.
func (c *Client) newRequestWithContext(ctx context.Context, method, urlStr string, body interface{}) (*http.Request, error) {
    u, err := c.BaseURL.Parse(strings.TrimPrefix(urlStr, "/"))
    if err != nil {
        return nil, err
    }

    var buf io.ReadWriter
    if body != nil {
        buf = new(bytes.Buffer)
        enc := json.NewEncoder(buf)
        enc.SetEscapeHTML(false)
        err := enc.Encode(body)
        if err != nil {
            return nil, err
        }
    }

    req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Accept", "application/json")
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", c.token)

    return req, nil
}

// ErrorResponse represents an error response from the Snipe-IT API.
//
// Snipe-IT API error responses typically contain a message explaining
// what went wrong, which is captured in the Message field.
type ErrorResponse struct {
    // Response is the HTTP response that generated the error
    Response *http.Response
    
    // Message is the error message returned by the Snipe-IT API
    Message  string `json:"message"`
}

// Error returns a string representation of the error.
// It implements the error interface.
func (e *ErrorResponse) Error() string {
    return fmt.Sprintf("%v %v: %d %v",
        e.Response.Request.Method, e.Response.Request.URL,
        e.Response.StatusCode, e.Message)
}

// Do sends an API request and returns the API response.
//
// req is the HTTP request to send.
// v is the destination into which the response JSON will be unmarshaled.
// If v implements the io.Writer interface, the raw response body will be written to it
// without attempting to parse it as JSON.
//
// If the response status code is not in the 2xx range, an ErrorResponse is returned.
// Otherwise, if v is not nil, the response body is JSON decoded into v.
//
// The provided request and returned response are for debugging purposes only and
// should not be directly modified.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
    return c.DoWithOptions(req, v, nil)
}

// DoContext sends an API request with the provided context and returns the API response.
//
// ctx is the context for the request. If it is nil, a default context will be used.
// req is the HTTP request to send.
// v is the destination into which the response JSON will be unmarshaled.
// If v implements the io.Writer interface, the raw response body will be written to it
// without attempting to parse it as JSON.
//
// If the response status code is not in the 2xx range, an ErrorResponse is returned.
// Otherwise, if v is not nil, the response body is JSON decoded into v.
//
// The provided request and returned response are for debugging purposes only and
// should not be directly modified.
func (c *Client) DoContext(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
    if ctx == nil {
        ctx = context.Background()
    }
    
    opts := &RequestOptions{Context: ctx}
    return c.DoWithOptions(req, v, opts)
}

// AddOptions adds the parameters in opt as URL query parameters to s.
//
// s is the URL string to which the query parameters will be added.
// opt is a struct whose fields contain "url" tags that define the query parameter names.
// 
// For example, if opt is a struct with a field `Limit int `url:"limit,omitempty"`,
// and Limit is set to 10, the resulting URL will have ?limit=10 appended.
//
// This method relies on the github.com/google/go-querystring package to
// convert the struct fields to query parameters.
func (c *Client) AddOptions(s string, opt interface{}) (string, error) {
    v := reflect.ValueOf(opt)
    if v.Kind() == reflect.Ptr && v.IsNil() {
        return s, nil
    }

    u, err := url.Parse(s)
    if err != nil {
        return s, err
    }

    qs, err := query.Values(opt)
    if err != nil {
        return s, err
    }

    u.RawQuery = qs.Encode()
    return u.String(), nil
}
