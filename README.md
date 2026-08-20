# go-snipeit

A Go client library for the Snipe-IT Asset Management API.

## Rate limiting

Snipe-IT advertises its per-window budget on every response
(`X-Ratelimit-Limit`, `X-Ratelimit-Remaining`, `X-Ratelimit-Reset`,
`X-Ratelimit-Reset-Timestamp`, and `Retry-After` on a 429). The client records
the latest snapshot and can pace itself from it:

```go
limiter := snipeit.NewAdaptiveRateLimiter(4, 5) // start at 4 req/s (240/min)

client, err := snipeit.NewClientWithOptions(baseURL, token, &snipeit.ClientOptions{
    RateLimiter: limiter,
    OnRateLimit: func(rl snipeit.RateLimit) {
        log.Printf("snipe-it budget: %d/%d left, resets in %s", rl.Remaining, rl.Limit, rl.Reset)
    },
})

rl := client.RateLimit() // most recent snapshot, RateLimit.Valid until first response
```

`AdaptiveRateLimiter` never runs faster than the rate you configure; it slows
down as `X-Ratelimit-Remaining` drains and holds requests until the window
resets once the budget is spent. `ParseRateLimit` is exported for callers that
issue their own HTTP requests alongside the client.

Presets cover the published Snipe-IT Cloud allowances, so you don't have to
convert per-minute figures yourself:

```go
preset, ok := snipeit.PresetByName(cfg.SnipeITPlan) // "basic" | "small_business" | "dedicated"
if !ok {
    preset = snipeit.PresetBasic
}

client, err := snipeit.NewClientWithOptions(baseURL, token, &snipeit.ClientOptions{
    RateLimiter: preset.Limiter(), // nil for "dedicated", i.e. no limiting
})
```

| Preset | Plan | Allowance |
| --- | --- | --- |
| `snipeit.PresetBasic` | Basic | 120 req/min (2 req/s) |
| `snipeit.PresetSmallBusiness` | Small Business | 240 req/min (4 req/s) |
| `snipeit.PresetDedicated` | Dedicated | unmetered |

A preset only sets the ceiling — the limiter it returns still tightens its pace
from the response headers.
