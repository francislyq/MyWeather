package model

import "time"

type City struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CountryID int       `json:"country_id"`
	Country   string    `json:"country"`
	Sunrise   time.Time `json:"sunrise"`
	Sunset    time.Time `json:"sunset"`

	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Timezone  int     `json:"timezone"`
}

type Weather struct {
	CityID   int    `json:"city_id"`
	CityName string `json:"city_name"`

	Temperature float64 `json:"temperature"`
	FeelsLike   float64 `json:"feels_like"`
	TempMin     float64 `json:"temp_min"`
	TempHigh    float64 `json:"temp_high"`
	Pressure    int     `json:"pressure"`
	Humidity    int     `json:"humidity"`
	SeaLevel    int     `json:"sea_level"`
	GrndLevel   int     `json:"grnd_level"`

	WeatherMain string `json:"weather_main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`

	Visibility int     `json:"visibility"`
	WindSpeed  float64 `json:"wind_speed"`
	WindDeg    int     `json:"wind_deg"`
	WindGust   float64 `json:"wind_gust"`

	Rain1h    float64 `json:"rain_1h"`
	CloudsAll int     `json:"clouds_all"`
}

type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type WeatherResult struct {
	Weather   *Weather  `json:"weather,omitempty"`
	CacheHit  bool      `json:"-"`
	IsStale   bool      `json:"-"`
	FetchedAt time.Time `json:"-"`
	ExpiresAt time.Time `json:"-"`
}

type CityWeatherResult struct {
	CityID    int      `json:"city_id"`
	CityName  string   `json:"city_name"`
	Weather   *Weather `json:"weather,omitempty"`
	Status    string   `json:"status"`
	Cache     string   `json:"cache"`
	LatencyMs int64    `json:"latency_ms"`
	Error     string   `json:"error,omitempty"`
}

type CacheStatsResponse struct {
	Hits   int `json:"hits"`
	Misses int `json:"misses"`
	Size   int `json:"size"`
}
