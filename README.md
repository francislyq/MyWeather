# MyWeather
A cache-aware weather aggregation API built in Go. Fetches weather data from OpenWeatherMap, caches results in memory with configurable TTLs, and handles concurrent requests efficiently with stampede prevention.

## Tech Stack

- **Go** with [Uber FX](https://github.com/uber-go/fx) for dependency injection
- **gorilla/mux** for HTTP routing
- **logrus** for structured JSON logging (Mezmo-compatible)
- **golang.org/x/sync/singleflight** for request coalescing
- **testify** for testing

## Quick Start

### Prerequisites

- Go 1.22+
- An OpenWeatherMap API key

### Setup

```bash
# Clone the repo
git clone https://github.com/yuqiliu/myweather.git
cd myweather

# Create your .env from the template
cp .env.example .env
# Edit .env and set your WEATHER_API_KEY

# Run
go run .
```

### Docker

```bash
docker build -t myweather .
docker run -p 3000:3000 -e WEATHER_API_KEY=your_key_here myweather
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/cities` | List all configured cities |
| GET | `/cities/{id}/weather` | Get current weather for a city (cache-aware) |
| POST | `/weather/collect` | Bulk fetch weather for all cities concurrently |
| GET | `/cache/stats` | Cache statistics (hits, misses, size) |
| DELETE | `/cache` | Flush the entire cache |
| GET | `/health` | Health check |

### Response Headers

Weather endpoints include cache metadata:

| Header | Description |
|--------|-------------|
| `X-Cache` | `HIT`, `STALE`, or `MISS` |
| `X-Cache-TTL-Remaining` | Seconds until the entry's fresh window expires (0 when stale) |
| `X-Cache-Fetched-At` | RFC3339 timestamp when data was fetched from the external API |

## Configuration

All configuration is via environment variables (loaded from `.env` in development):

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `3000` | HTTP server port |
| `WEATHER_API_KEY` | *required* | OpenWeatherMap API key |
| `HTTP_TIMEOUT` | `10s` | External HTTP client timeout |
| `CACHE_TTL` | `5m` | Cache entry fresh window (time-to-live before becoming stale) |
| `CACHE_CLEANUP_INTERVAL` | `1m` | Background expired entry cleanup interval |
| `CACHE_STALE_WHILE_REVALIDATE` | `1m` | How long a stale entry may be served while a background refresh runs. Set to `0` to disable stale serving and revert to strict TTL behaviour. |

### GitHub Secrets (for CI/CD)

Store `WEATHER_API_KEY` as a GitHub repository secret. It will be injected as an environment variable in Actions/deployments.

## Cache Design

### Architecture

The cache is implemented as a standalone package (`internal/cache/`) with a clean interface that could be swapped for Redis or any other backing store.

### Entry Lifecycle

Each cache entry has two deadlines:

```
 stored          fresh window ends     stale window ends
   │                    │                     │
   ▼                    ▼                     ▼
───●────────────────────●─────────────────────●────▶ time
   │◄── CACHE_TTL ─────►│◄─ STALE_WHILE_REV ─►│
        (fresh, HIT)        (stale, STALE)       (expired, MISS)
```

- **Fresh** (`now < ExpiresAt`): served directly, `X-Cache: HIT`.
- **Stale** (`ExpiresAt ≤ now < StaleUntil`): stale data returned immediately (`X-Cache: STALE`) while a background goroutine refreshes the cache. The next request will get fresh data.
- **Expired** (`now ≥ StaleUntil`): entry evicted, treated as a cache miss — the request blocks on an upstream API call.

### Key Decisions

- **Stale-while-revalidate**: Stale data is served for up to `CACHE_STALE_WHILE_REVALIDATE` after the TTL expires, while a background goroutine silently refreshes the cache. This eliminates the latency spike that would otherwise hit every first request after a TTL expiry.
- **Configurable feature flag**: Setting `CACHE_STALE_WHILE_REVALIDATE=0` disables stale serving entirely, restoring strict TTL behaviour for contexts where data freshness is more important than latency consistency.
- **Background revalidation via singleflight**: The background refresh is wrapped in a `singleflight.Group` (keyed separately from foreground requests). If multiple goroutines detect a stale entry at the same time, only one upstream call is made.
- **Concurrent safety**: `sync.RWMutex` protects the map — read-lock for gets, write-lock for sets/deletes. Stats use `sync/atomic` counters to avoid lock contention on hot paths.
- **Stampede prevention on cold misses**: `singleflight` in the weather service ensures that concurrent requests for the same fully-expired city result in exactly one external API call. Other goroutines wait and share the result.
- **Periodic eviction**: A background goroutine sweeps the map on `CACHE_CLEANUP_INTERVAL` and removes any entries whose stale window has also expired. Lazy eviction on `Get()` handles the rest.

### Trade-offs

| Concern | Decision | Rationale |
|---------|----------|-----------|
| Latency consistency | Serve stale, revalidate in background | Eliminates the "cold cache spike" after every TTL expiry; weather data is tolerant of being seconds or a minute old |
| Data freshness | Configurable stale window (default 1 m) | Short stale window keeps data reasonably fresh while absorbing TTL boundary latency |
| Complexity | Added `staleUntil` field and background goroutine | Small increase; the singleflight guard prevents background stampede |
| Stale data risk | `X-Cache: STALE` header exposed to clients | Callers that require strict freshness can observe the header and decide to wait or retry |
| Memory | Entries live longer (TTL + stale window) | With 3 cities this is negligible; periodic + lazy eviction bounds growth |
| Cache-aside correctness | Background revalidation uses `context.Background()` | Foreground request context may cancel before revalidation finishes; a detached context prevents premature cancellation |

- **No LRU/max-size eviction**: With only 3 configured cities, memory is not a concern. The cache interface supports adding eviction policies later.
- **In-memory only**: Cache is lost on restart. Acceptable for this use case where data is cheap to re-fetch.

## Configured Cities

| City ID | Name | Country |
|---------|------|---------|
| 6167865 | Toronto | CA |
| 6094817 | Ottawa | CA |
| 1850147 | Tokyo | JP |

## Testing

```bash
# Run all tests with race detector
go test -race -v ./...

# Run benchmarks
go test -bench=. -benchmem ./internal/cache/

```

## Project Structure

```
main.go          # FX app wiring, entry point
internal/
  config/                   # Configuration loading, logger setup, FX module
  cache/                    # Cache interface + in-memory implementation
  weather/                  # Provider interface, OpenWeatherMap client, orchestration service
  handler/                  # HTTP handlers
  model/                    # Shared types
  server/                   # HTTP server lifecycle (FX-managed)
```
