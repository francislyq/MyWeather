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
| `X-Cache` | `HIT` or `MISS` |
| `X-Cache-TTL-Remaining` | Seconds until this entry expires |
| `X-Cache-Fetched-At` | RFC3339 timestamp when data was fetched from the external API |

## Configuration

All configuration is via environment variables (loaded from `.env` in development):

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `3000` | HTTP server port |
| `WEATHER_API_KEY` | *required* | OpenWeatherMap API key |
| `HTTP_TIMEOUT` | `10s` | External HTTP client timeout |
| `CACHE_TTL` | `5m` | Cache entry time-to-live |
| `CACHE_CLEANUP_INTERVAL` | `1m` | Background expired entry cleanup interval |

### GitHub Secrets (for CI/CD)

Store `WEATHER_API_KEY` as a GitHub repository secret. It will be injected as an environment variable in Actions/deployments.

## Cache Design

### Architecture

The cache is implemented as a standalone package (`internal/cache/`) with a clean interface that could be swapped for Redis or any other backing store.

### Key Decisions

- **Strict TTL enforcement**: Expired entries are never served. On `Get()`, if the entry is expired, it is deleted and a cache miss is returned.
- **Concurrent safety**: `sync.RWMutex` protects the map — read-lock for gets, write-lock for sets/deletes. Stats use `sync/atomic` counters to avoid lock contention on hot paths.
- **Stampede prevention**: `singleflight` in the weather service ensures that concurrent requests for the same uncached city result in exactly one external API call. Other goroutines wait and share the result.
- **Lazy + periodic eviction**: Expired entries are cleaned on access (lazy) and by a background goroutine on a configurable interval (periodic sweep).

### Trade-offs

- **No stale serving**: I chose strict TTL over stale-while-revalidate for simplicity and predictability. The trade-off is that the first request after expiry pays the full latency of an external API call.
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
